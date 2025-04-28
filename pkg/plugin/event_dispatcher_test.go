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
	crand "crypto/rand"
	"fmt"
	mathrand "math/rand"
	"net"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var (
	expectData []byte
	writevData [][]byte
	done       = make(chan struct{})
)

type serverConnCallback struct {
	t          *testing.T
	readBuffer []byte
}

type clientConnCallback struct {
	t *testing.T
}

func (c *clientConnCallback) onEventData(buf []byte, conn eventConn) error { return nil }
func (c *clientConnCallback) onRemoteClose()                               { fmt.Println("client onRemoteClose") }
func (c *clientConnCallback) onLocalClose()                                { fmt.Println("client onLocalClose") }

func (c *serverConnCallback) onEventData(buf []byte, conn eventConn) error {
	c.readBuffer = append(c.readBuffer, buf...)
	conn.commitRead(len(buf))
	time.Sleep(time.Millisecond)
	if len(c.readBuffer) == len(expectData) {
		//fmt.Println("c.readBufferLen", len(c.readBuffer), "expectData len", len(expectData))
		assert.Equal(c.t, c.readBuffer, expectData)
		close(done)
	}
	//fmt.Println("c.readBufferLen", len(c.readBuffer), "expectData len", len(expectData))
	return nil
}

func (c *serverConnCallback) onRemoteClose() { fmt.Println("server onRemoteClose") }
func (c *serverConnCallback) onLocalClose()  { fmt.Println("server onLocalClose") }

var _ eventConnCallback = &serverConnCallback{}
var _ eventConnCallback = &clientConnCallback{}

func fillTestingData() {
	const msgN = 1020
	writevData = make([][]byte, msgN)
	expectData = make([]byte, 0, 1024*1024*msgN)
	for i := 0; i < msgN; i++ {
		writevData[i] = make([]byte, mathrand.Intn(1*1024*1024))
		_, _ = crand.Read(writevData[i])
		expectData = append(expectData, writevData[i]...)
		//fmt.Println("slice i ", i, len(writevData[i]))
	}
}

type EventDispatcherTestSuite struct {
	suite.Suite
}

func (s *EventDispatcherTestSuite) TestEventDispatcher() {
	ensureDefaultDispatcherInit()
	fillTestingData()
	d := defaultDispatcher
	var clientConn, serverConn eventConn
	done := make(chan struct{})
	go func() {
		ln, err := net.Listen("tcp", ":7777")
		s.Require().Nil(err)
		for {
			conn, err := ln.Accept()
			if err != nil {
				fmt.Println(err)
				return
			}
			fd, err := getConnDupFd(conn)
			s.Require().Nil(err)
			defer func() {
				if err := conn.Close(); err != nil {
					s.T().Fatalf("conn.Close failed: %v", err)
				}
			}()
			serverConn = d.newConnection(fd)
			s.Require().Nil(serverConn.setCallback(&serverConnCallback{
				t:          s.T(),
				readBuffer: make([]byte, 0, len(expectData)),
			}))
			runtime.KeepAlive(fd)
		}
	}()

	time.Sleep(100 * time.Millisecond)

	go func() {
		conn, err := net.Dial("tcp", ":7777")
		s.Require().Nil(err)
		fd, err := getConnDupFd(conn)
		s.Require().Nil(err)
		defer func() {
			if err := conn.Close(); err != nil {
				s.T().Fatalf("conn.Close failed: %v", err)
			}
		}()
		clientConn = d.newConnection(fd)
		s.Require().Nil(clientConn.setCallback(&clientConnCallback{s.T()}))
		err = clientConn.write(writevData[0])
		s.Require().Nil(err)
		err = clientConn.writev(writevData[1:]...)
		s.Require().Nil(err)
		runtime.KeepAlive(fd)
	}()

	<-done
	clientConn.close() //nolint:errcheck // test cleanup, error ignored intentionally
	serverConn.close() //nolint:errcheck // test cleanup, error ignored intentionally
}

func TestEventDispatcherTestSuite(t *testing.T) {
	suite.Run(t, new(EventDispatcherTestSuite))
}
