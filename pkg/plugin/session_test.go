/*
 * Copyright 2025 SREDiag Authors
 * Copyright 2023 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package plugin

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type SessionTestSuite struct {
	suite.Suite
}

func testConn() (*net.UnixConn, *net.UnixConn) {
	return testUdsConn()
}

func testUdsConn() (client *net.UnixConn, server *net.UnixConn) {
	udsPath := "shmipc_unit_test" + strconv.Itoa(int(rand.Int63())) + "_" + strconv.Itoa(time.Now().Nanosecond())
	_ = syscall.Unlink(udsPath)
	addr := &net.UnixAddr{Net: "unix", Name: udsPath}
	notifyCh := make(chan struct{})
	go func() {
		defer func() {
			_ = syscall.Unlink(udsPath)
		}()
		ln, err := net.ListenUnix("unix", addr)
		if err != nil {
			panic("create listener failed:" + err.Error())
		}
		notifyCh <- struct{}{}
		server, err = ln.AcceptUnix()
		if err != nil {
			panic("accept conn failed:" + err.Error())
		}
		_ = ln.Close() // Ignore error on uds listener close in test setup
		notifyCh <- struct{}{}
	}()
	<-notifyCh
	var err error
	client, err = net.DialUnix("unix", nil, addr)
	if err != nil {
		panic("dial uds failed:" + err.Error())
	}
	<-notifyCh
	return
}

func testConf() *Config {
	conf := DefaultConfig()
	conf.MemMapType = MemMapTypeMemFd
	conf.ConnectionWriteTimeout = 25000 * time.Millisecond
	conf.ShareMemoryPathPrefix = "/Volumes/RAMDisk/plugin.test"
	if runtime.GOOS == "linux" {
		conf.ShareMemoryPathPrefix = "/dev/shm/plugin.test_" + strconv.Itoa(int(rand.Int63()))
	}
	conf.ShareMemoryBufferCap = 32 * 1024 * 1024 // 32M
	conf.LogOutput = os.Stdout
	return conf
}

func testClientServer() (*Session, *Session) {
	return testClientServerConfig(testConf())
}

func testClientServerConfig(conf *Config) (*Session, *Session) {
	clientConn, serverConn := testConn()
	var server *Session
	ok := make(chan struct{})

	go func() {
		var sErr error
		serverConf := *conf
		server, sErr = newSession(&serverConf, serverConn, false)
		if sErr != nil {
			panic(sErr)
		}
		close(ok)
	}()
	clientConf := *conf
	client, cErr := newSession(&clientConf, clientConn, true)
	if cErr != nil {
		panic(cErr)
	}
	<-ok
	return client, server
}

func (s *SessionTestSuite) TestSession_OpenStream() {
	fmt.Println("----------test session open stream----------")
	client, server := testClientServer()

	// case1: session closed
	_ = client.Close()
	_ = server.Close()
	s.Require().True(client.IsClosed())
	stream, err := client.OpenStream()
	s.Require().Nil(stream)
	s.Require().Equal(ErrSessionShutdown, err)
	// client's Close will cause server's Close
	s.Require().True(server.IsClosed())

	//case2: CircuitBreaker triggered
	client2, server2 := testClientServer()
	client2.openCircuitBreaker()
	stream2, err := client2.OpenStream()
	s.Require().Nil(stream2)
	s.Require().Equal(ErrSessionUnhealthy, err)
	if err := client2.Close(); err != nil {
		s.T().Logf("client2.Close error: %v", err)
	}
	if err := server2.Close(); err != nil {
		s.T().Logf("server2.Close error: %v", err)
	}

	// case3: stream exist
	client3, server3 := testClientServer()
	client3.streams.Set(strconv.Itoa(int(client3.nextStreamID+1)), newStream(client3, client3.nextStreamID+1))
	stream3, err := client3.OpenStream()
	s.Require().Nil(stream3)
	s.Require().Equal(ErrStreamsExhausted, err)
	if err := client3.Close(); err != nil {
		s.T().Logf("client3.Close error: %v", err)
	}
	if err := server3.Close(); err != nil {
		s.T().Logf("server3.Close error: %v", err)
	}
}

func (s *SessionTestSuite) TestSession_AcceptStreamNormally() {
	fmt.Println("----------test session accept stream normally----------")
	done := make(chan struct{})
	notifyRead := make(chan struct{})
	client, server := testClientServer()
	defer func() {
		if err := client.Close(); err != nil {
			s.T().Logf("client.Close error: %v", err)
		}
		if err := server.Close(); err != nil {
			s.T().Logf("server.Close error: %v", err)
		}
	}()
	// case1: accept stream normally
	go func() {
		cStream, err := client.OpenStream()
		defer func() {
			if err := cStream.Close(); err != nil {
				s.T().Logf("cStream.Close error: %v", err)
			}
		}()
		if err != nil {
			s.FailNow("Failed to open stream:", err.Error())
		}
		// only when we actually send something, the server can
		// aware that a new stream created, therefore we need to
		// send a byte to notify server
		_ = cStream.BufferWriter().WriteString("1")
		if err := cStream.Flush(true); err != nil {
			s.T().Logf("cStream.Flush error: %v", err)
		}
		// wait resp
		<-notifyRead

		respData, err := cStream.BufferReader().ReadBytes(1)
		if err != nil {
			s.FailNow("Failed to read bytes:", err.Error())
		}
		s.Require().Equal("1", string(respData))
		close(done)
	}()

	sStream, err := server.AcceptStream()
	if err != nil {
		s.FailNow("Failed to accept stream:", err.Error())
	}
	defer func() {
		if err := sStream.Close(); err != nil {
			s.T().Logf("sStream.Close error: %v", err)
		}
	}()
	respData, err := sStream.BufferReader().ReadBytes(1)
	if err != nil {
		s.FailNow("Failed to read bytes:", err.Error())
	}
	s.Require().Equal("1", string(respData))

	// write back
	_ = sStream.BufferWriter().WriteString("1")
	if err := sStream.Flush(true); err != nil {
		s.T().Logf("sStream.Flush error: %v", err)
	}
	close(notifyRead)
	<-done
	fmt.Println("----------test session accept stream normally done----------")
}

func (s *SessionTestSuite) TestSession_AcceptStreamWhenSessionClosed() {
	fmt.Println("----------test session accept stream when session closed----------")
	client, server := testClientServer()
	defer func() {
		if err := client.Close(); err != nil {
			s.T().Logf("client.Close error: %v", err)
		}
		if err := server.Close(); err != nil {
			s.T().Logf("server.Close error: %v", err)
		}
	}()
	done := make(chan struct{})
	notifyAccept := make(chan struct{})
	// case 2: accept when session closed
	go func() {
		<-notifyAccept
		// try accept after session shutdown
		_, err := server.AcceptStream()
		s.Require().Error(ErrSessionShutdown, err)
		close(done)
	}()

	cStream2, err := client.OpenStream()
	defer func() {
		if err := cStream2.Close(); err != nil {
			s.T().Logf("cStream2.Close error: %v", err)
		}
	}()

	if err != nil {
		s.FailNow("Failed to malloc buf:", err.Error())
	}
	_ = cStream2.BufferWriter().WriteString("1")
	if err := cStream2.Flush(true); err != nil {
		s.T().Logf("cStream2.Flush error: %v", err)
	}

	// now shutdown session
	_ = client.Close()
	_ = server.Close()
	s.Require().True(server.shutdown == 1)
	close(notifyAccept)
	<-done
	fmt.Println("----------test session accept stream when session closed done----------")
}

func (s *SessionTestSuite) TestSendData_Small() {
	client, server := testClientServer()
	defer func() {
		if err := client.Close(); err != nil {
			s.T().Logf("client.Close error: %v", err)
		}
		if err := server.Close(); err != nil {
			s.T().Logf("server.Close error: %v", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		stream, err := server.AcceptStream()
		if err != nil {
			s.T().Logf("err: %v", err)
		}

		if server.GetActiveStreamCount() != 1 {
			s.Fail("num of streams is ", server.GetActiveStreamCount())
		}

		size := 0
		for size < 4*100 {
			bs, err := stream.BufferReader().ReadBytes(4)
			size += 4
			if err != nil {
				s.T().Errorf("read err: %v", err)
				return
			}
			if string(bs) != "test" {
				s.T().Logf("bad: %s", string(bs))
			}
		}

		if err := stream.Close(); err != nil {
			s.T().Logf("err: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		stream, err := client.OpenStream()
		if err != nil {
			s.T().Logf("err: %v", err)
		}

		if client.GetActiveStreamCount() != 1 {
			s.T().Logf("bad")
		}

		for i := 0; i < 100; i++ {
			_, err := stream.BufferWriter().WriteBytes([]byte("test"))
			if err != nil {
				s.T().Logf("err: %v", err)
			}
			err = stream.Flush(false)
			if err != nil {
				s.T().Logf("stream.Flush error: %v", err)
			}
		}

		if err := stream.Close(); err != nil {
			s.T().Logf("err: %v", err)
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second * 35):
		panic("timeout")
	}

	if client.GetActiveStreamCount() != 0 {
		s.T().Fatalf("bad, streams:%d", client.GetActiveStreamCount())
	}
	if server.GetActiveStreamCount() != 0 {
		s.T().Fatalf("bad")
	}
}

func TestSendData_Large(t *testing.T) {
	client, server := testClientServer()
	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("client.Close error: %v", err)
		}
		if err := server.Close(); err != nil {
			t.Logf("server.Close error: %v", err)
		}
	}()

	const (
		sendSize = 2 * 1024 * 1024
		recvSize = 4 * 1024
	)

	data := make([]byte, recvSize)
	for idx := range data {
		data[idx] = byte(idx % 256)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		stream, err := server.AcceptStream()
		if err != nil {
			t.Logf("err: %v", err)
		}
		defer func() {
			if err := stream.Close(); err != nil {
				t.Logf("stream.Close error: %v", err)
			}
		}()
		for hadRead := 0; hadRead < sendSize; hadRead++ {
			if bt, err := stream.BufferReader().ReadByte(); bt != byte(hadRead%256) || err != nil {
				t.Logf("bad: %v %v", hadRead, bt)
			}

		}
	}()

	go func() {
		defer wg.Done()
		stream, err := client.OpenStream()
		if err != nil {
			t.Logf("err: %v", err)
		}

		for i := 0; i < sendSize/recvSize; i++ {
			_, _ = stream.BufferWriter().WriteBytes(data)
			err = stream.Flush(true)
			if err != nil {
				t.Logf("err: %v", err)
			}
		}

		if err := stream.Close(); err != nil {
			t.Logf("err: %v", err)
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		panic("timeout")
	}
}

func TestManyStreams(t *testing.T) {
	client, server := testClientServer()
	defer func() {
		if err := server.Close(); err != nil {
			t.Fatalf("server.Close error: %v", err)
		}
		if err := client.Close(); err != nil {
			t.Fatalf("client.Close error: %v", err)
		}
	}()

	wg := &sync.WaitGroup{}
	const writeSize = 8
	acceptor := func(i int) {
		defer wg.Done()
		stream, err := server.AcceptStream()
		if err != nil {
			t.Errorf("AcceptStream[%d] err: %v", i, err)
			return
		}
		defer func() {
			if err := stream.Close(); err != nil {
				t.Logf("stream.Close error: %v", err)
			}
		}()

		for {
			if err := stream.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
				t.Errorf("SetReadDeadline[%d] error: %v", i, err)
				return
			}
			buf := stream.BufferReader()
			n, err := buf.Discard(writeSize)
			if err == ErrEndOfStream {
				return
			}
			if err == io.EOF || err == ErrTimeout {
				// t.Logf("ret err:%s", err)
				return
			}
			if err != nil {
				t.Errorf("Discard[%d] err: %v", i, err)
				return
			}
			if buf.Len() != 0 {
				t.Errorf("Discard[%d] buf.Len() != 0: %d n:%d", i, buf.Len(), n)
				return
			}
		}
	}
	sender := func(i int) {
		defer wg.Done()
		stream, err := client.OpenStream()
		if err != nil {
			t.Errorf("OpenStream[%d] err: %v", i, err)
			return
		}
		defer func() {
			if err := stream.Close(); err != nil {
				t.Logf("stream.Close error: %v", err)
			}
		}()

		var msg [writeSize]byte
		if _, err := stream.BufferWriter().WriteBytes(msg[:]); err != nil {
			t.Errorf("err: %v", err)
			return
		}
		err = stream.Flush(true)
		if err != nil {
			t.Errorf("Flush[%d] err: %v", i, err)
			return
		}
	}

	for i := 0; i < 1; i++ {
		wg.Add(2)
		go acceptor(i)
		go sender(i)
	}

	wg.Wait()
}

func TestSendData_Small_Memfd(t *testing.T) {
	conf := testConf()
	conf.MemMapType = MemMapTypeMemFd
	client, server := testClientServer()

	defer func() {
		if err := client.Close(); err != nil {
			t.Logf("client.Close error: %v", err)
		}
		if err := server.Close(); err != nil {
			t.Logf("server.Close error: %v", err)
		}
	}()
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		stream, err := server.AcceptStream()
		if err != nil {
			t.Logf("err: %v", err)
		}

		if server.GetActiveStreamCount() != 1 {
			t.Errorf("Unexpected stream count: %d", server.GetActiveStreamCount())
			return
		}

		size := 0
		for size < 4*100 {
			bs, err := stream.BufferReader().ReadBytes(4)
			size += 4
			if err != nil {
				t.Errorf("ReadBytes err: %v", err)
				return
			}
			if string(bs) != "test" {
				t.Logf("bad: %s", string(bs))
			}
		}

		if err := stream.Close(); err != nil {
			t.Logf("err: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		stream, err := client.OpenStream()
		if err != nil {
			t.Logf("err: %v", err)
		}

		if client.GetActiveStreamCount() != 1 {
			t.Errorf("Unexpected stream count: %d", client.GetActiveStreamCount())
			return
		}

		for i := 0; i < 100; i++ {
			_, err := stream.BufferWriter().WriteBytes([]byte("test"))
			if err != nil {
				t.Logf("err: %v", err)
			}
			err = stream.Flush(false)
			if err != nil {
				t.Logf("stream.Flush error: %v", err)
			}
		}

		if err := stream.Close(); err != nil {
			t.Logf("err: %v", err)
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		wg.Wait()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(time.Second * 35):
		panic("timeout")
	}

	if client.GetActiveStreamCount() != 0 {
		t.Fatalf("bad, streams:%d", client.GetActiveStreamCount())
	}
	if server.GetActiveStreamCount() != 0 {
		t.Fatalf("bad")
	}
}
