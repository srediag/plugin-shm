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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type StreamTestSuite struct {
	suite.Suite
}

// T returns the underlying *testing.T for use in assertions and fatal calls.
func (suite *StreamTestSuite) T() *testing.T {
	return suite.Suite.T()
}

func newClientServerWithNoCheck(conf *Config) (client *Session, server *Session) {
	conn1, conn2 := testConn()

	serverOk := make(chan struct{})
	go func() {
		var err error
		server, err = newSession(conf, conn2, false)
		if err != nil {
			panic(err)
		}
		close(serverOk)
	}()

	cconf := testConf()
	*cconf = *conf

	var err error
	client, err = newSession(cconf, conn1, true)
	if err != nil {
		panic(err)
	}
	<-serverOk
	return client, server
}

// TestStream_Close verifies that closing a stream transitions through the correct states
// and that the server's active stream count is updated as expected. It checks the sequence
// of state transitions and ensures that resources are released properly after closing.
func (suite *StreamTestSuite) TestStream_Close() {
	conf := testConf()
	conf.QueueCap = 8
	client, server := newClientServerWithNoCheck(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	hadForceCloseNotifyCh := make(chan struct{})
	doneCh := make(chan struct{})

	go func() {
		strm, err := server.AcceptStream()
		if err != nil {
			suite.T().Errorf("accept stream failed: %s", err.Error())
			return
		}
		_, err = strm.BufferReader().ReadByte()
		if err != nil {
			suite.Require().Equal(nil, err)
		}
		<-hadForceCloseNotifyCh
		_, err = strm.BufferReader().ReadByte()
		if err != nil {
			suite.Require().Equal(ErrEndOfStream, err)
		}
		suite.Require().Equal(1, server.GetActiveStreamCount())
		suite.Require().Equal(uint32(streamHalfClosed), strm.state)
		suite.Require().Equal(0, server.GetActiveStreamCount())
		suite.Require().Equal(uint32(streamClosed), strm.state)
		suite.Require().Equal(1, server.GetActiveStreamCount())
		suite.Require().Equal(uint32(streamHalfClosed), strm.state)
		suite.Require().Equal(0, server.GetActiveStreamCount())
		suite.Require().Equal(uint32(streamClosed), strm.state)
		suite.Require().Equal(0, server.GetActiveStreamCount())
		close(doneCh)
	}()
	stream, err := client.OpenStream()
	if err != nil {
		suite.T().Fatalf("open stream failed:,err:%s", err.Error())
	}
	_ = stream.BufferWriter().WriteString("1")
	_ = stream.Flush(false)
	close(hadForceCloseNotifyCh)
	<-doneCh
}

// TestStream_ClientFallback verifies that when the client falls back from shared memory to another transport,
// data is still transmitted correctly and the stream behaves as expected.
func (suite *StreamTestSuite) TestStream_ClientFallback() {
	conf := testConf()
	conf.ShareMemoryBufferCap = 1 << 20
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()
	done := make(chan struct{})
	errCh := make(chan error, 1)
	mockData := make([][]byte, 2000)
	dataSize := 1024
	for i := range mockData {
		mockData[i] = make([]byte, dataSize)
		if _, err := crand.Read(mockData[i]); err != nil {
			suite.T().Fatalf("rand.Read failed: %v", err)
		}
	}
	go func() {
		stream, err := server.AcceptStream()
		if err != nil {
			errCh <- err
			return
		}
		defer func() { _ = stream.Close() }()
		reader := stream.BufferReader()
		for i := range mockData {
			get, err := reader.ReadBytes(dataSize)
			if err != nil {
				errCh <- err
				return
			}
			if !assert.Equal(suite.T(), mockData[i], get, "i:%d", i) {
				errCh <- err
				return
			}
		}
		close(done)
	}()
	stream, err := client.OpenStream()
	if err != nil {
		suite.T().Fatalf("client.OpenStream failed:%s", err.Error())
	}
	defer func() { _ = stream.Close() }()
	writer := stream.BufferWriter()
	for i := range mockData {
		if _, err = writer.WriteBytes(mockData[i]); err != nil {
			suite.T().Fatalf("Malloc(dataSize) failed:%s", err.Error())
		}
	}
	if lb, ok := writer.(*linkedBuffer); ok {
		assert.Equal(suite.T(), false, lb.isFromShareMemory())
	}
	err = stream.Flush(true)
	if err != nil {
		suite.T().Fatalf("writeBuf failed:%s", err.Error())
	}
	select {
	case <-done:
		// success
	case err := <-errCh:
		if err != nil {
			suite.T().Fatalf("goroutine error: %v", err)
		}
	}
}

// TestStream_ServerFallback verifies that when the server falls back from shared memory to another transport,
// data is still transmitted correctly and the stream behaves as expected.
func (suite *StreamTestSuite) TestStream_ServerFallback() {
	conf := testConf()
	conf.ShareMemoryBufferCap = 1 << 20
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()
	done := make(chan struct{})
	errCh := make(chan error, 1)
	mockData := make([][]byte, 2000)
	dataSize := 1024
	for i := range mockData {
		mockData[i] = make([]byte, dataSize)
		if _, err := crand.Read(mockData[i]); err != nil {
			suite.T().Fatalf("rand.Read failed: %v", err)
		}
	}
	go func() {
		stream, err := server.AcceptStream()
		if err != nil {
			errCh <- err
			return
		}
		defer func() { _ = stream.Close() }()
		writer := stream.BufferWriter()
		for i := range mockData {
			if _, err = writer.WriteBytes(mockData[i]); err != nil {
				errCh <- err
				return
			}
		}
		if lb, ok := writer.(*linkedBuffer); ok {
			assert.Equal(suite.T(), false, lb.isFromShareMemory())
		}
		err = stream.Flush(true)
		if err != nil {
			errCh <- err
			return
		}
		close(done)
	}()
	stream, err := client.OpenStream()
	if err != nil {
		suite.T().Fatalf("client.OpenStream failed:%s", err.Error())
	}
	writer := stream.BufferWriter()
	if _, err = writer.WriteBytes(mockData[0]); err != nil {
		suite.T().Fatalf("Malloc(dataSize) failed:%s", err.Error())
	}
	err = stream.Flush(true)
	if err != nil {
		suite.T().Fatalf("writeBuf failed:%s", err.Error())
	}
	reader := stream.BufferReader()
	for i := range mockData {
		get, err := reader.ReadBytes(dataSize)
		if err != nil {
			suite.T().Fatalf("ReadBytes failed:%s", err.Error())
		}
		assert.Equal(suite.T(), mockData[i], get, "i:%d", i)
	}
	defer func() { _ = stream.Close() }()
	select {
	case <-done:
		// success
	case err := <-errCh:
		if err != nil {
			suite.T().Fatalf("goroutine error: %v", err)
		}
	}
}

// TestStream_RandomPackageSize is not yet implemented.
// TODO: Implement this test or remove if not needed.

// TestStream_HalfClose is not yet implemented.
// TODO: Implement this test or remove if not needed.

func (suite *StreamTestSuite) TestStream_SendQueueFull() {
	conf := testConf()
	conf.QueueCap = 8 // can only contain 1 queue elem(12B)
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	done := make(chan struct{})
	errCh := make(chan error, 1)
	dataSize := 10
	mockDataLength := 500
	mockData := make([][]byte, mockDataLength)
	for i := range mockData {
		mockData[i] = make([]byte, dataSize)
		mockData[i][0] = byte(i)
	}

	stream, err := client.OpenStream()
	if err != nil {
		suite.T().Fatalf("client.OpenStream failed:%s", err.Error())
	}
	defer func() { _ = stream.Close() }()
	for i := 0; i < mockDataLength; i++ {
		writer := stream.BufferWriter()
		if _, err = writer.WriteBytes(mockData[i]); err != nil {
			suite.T().Fatalf("Malloc(dataSize) failed:%s", err.Error())
		}
		err = stream.Flush(true)
		if err != nil {
			suite.T().Fatalf("writeBuf failed:%s", err.Error())
		}
	}

	go func() {
		stream, err := server.AcceptStream()
		if err != nil {
			errCh <- err
			return
		}
		defer func() { _ = stream.Close() }()
		for i := 0; i < mockDataLength; i++ {
			reader := stream.BufferReader()
			get, err := reader.ReadBytes(dataSize)
			if err != nil {
				errCh <- err
				return
			}
			if !assert.Equal(suite.T(), mockData[i], get, "i:%d", i) {
				errCh <- err
				return
			}
		}
		close(done)
	}()

	select {
	case <-done:
		// success
	case err := <-errCh:
		if err != nil {
			suite.T().Fatalf("goroutine error: %v", err)
		}
	}
}

func (suite *StreamTestSuite) TestStream_SendQueueFullTimeout() {
	conf := testConf()
	conf.QueueCap = 8 // can only contain 1 queue elem(12B)
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	dataSize := 10
	mockDataLength := 500
	mockData := make([][]byte, mockDataLength)
	for i := range mockData {
		mockData[i] = make([]byte, dataSize)
		if _, err := crand.Read(mockData[i]); err != nil {
			suite.T().Fatalf("rand.Read failed: %v", err)
		}
	}

	stream, err := client.OpenStream()
	// if send queue is full, it will trigger timeout immediately
	_ = stream.SetWriteDeadline(time.Now())
	if err != nil {
		suite.T().Fatalf("client.OpenStream failed:%s", err.Error())
	}
	defer func() { _ = stream.Close() }()

	// write in a short time to full the send queue
	for i := 0; i < mockDataLength; i++ {
		writer := stream.BufferWriter()
		if _, err = writer.WriteBytes(mockData[i]); err != nil {
			suite.T().Fatalf("Malloc(dataSize) failed:%s", err.Error())
		}
		err = stream.Flush(true)
		if err != nil {
			// must be timeout err here
			assert.Error(suite.T(), ErrTimeout, err)
			// reset deadline
			if err := stream.SetDeadline(time.Now().Add(time.Microsecond * 10)); err != nil {
				suite.T().Fatalf("stream.SetDeadline error: %v", err)
			}
		}
	}
}

// reset                           92.9%
func (suite *StreamTestSuite) TestStream_Reset() {
	conf := testConf()
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	done := make(chan struct{})
	notifySend := make(chan struct{})
	notifyRead := make(chan struct{})
	notifyClosed := make(chan struct{})
	// use linked buffer by the way
	dataSize := 64<<10 + 1
	mockDataLength := 4
	mockData := make([][]byte, mockDataLength)
	for i := range mockData {
		mockData[i] = make([]byte, dataSize)
		if _, err := crand.Read(mockData[i]); err != nil {
			suite.T().Fatalf("rand.Read failed: %v", err)
		}
	}

	errCh := make(chan error, 1)
	go func() {
		s, err := server.AcceptStream()
		if err != nil {
			errCh <- err
			return
		}
		defer func() { _ = s.Close() }()
		reader := s.BufferReader()
		for i := 0; i < mockDataLength/2; i++ {
			get, err := reader.ReadBytes(dataSize)
			if err != nil {
				errCh <- err
				return
			}
			if !assert.Equal(suite.T(), mockData[i], get, "i:%d", i) {
				errCh <- err
				return
			}
		}
		// read done, reset will be successful
		err = s.reset()
		if err != nil {
			errCh <- err
			return
		}
		close(notifySend)
		<-notifyRead

		// test reuse
		reader = s.BufferReader()
		for i := mockDataLength / 2; i < mockDataLength; i++ {
			get, err := reader.ReadBytes(dataSize)
			if err != nil {
				errCh <- err
				return
			}
			if !assert.Equal(suite.T(), mockData[i], get, "i:%d", i) {
				errCh <- err
				return
			}
		}
		close(notifyRead)
		if err := s.Close(); err != nil {
			errCh <- fmt.Errorf("server stream s.Close error: %w", err)
		}
		close(notifyClosed)
		select {
		case err := <-errCh:
			if err != nil {
				errCh <- err
			}
		case <-done:
		}
	}()

	s, err := client.OpenStream()
	if err != nil {
		suite.T().Fatalf("client.OpenStream failed:%s", err.Error())
	}
	defer func() { _ = s.Close() }()

	for i := 0; i < mockDataLength; i++ {
		if i == mockDataLength/2 {
			<-notifySend
		}
		if _, err := s.BufferWriter().WriteBytes(mockData[i]); err != nil {
			suite.T().Fatalf("Malloc(dataSize) failed:%s", err.Error())
		}
		err = s.Flush(false)
		if err != nil {
			suite.T().Fatalf("writeBuf failed:%s", err.Error())
		}
	}
	close(notifyRead)
	if err := s.Close(); err != nil {
		suite.T().Logf("client stream s.Close error: %v", err)
	}
	close(notifyClosed)
	select {
	case err := <-errCh:
		if err != nil {
			suite.T().Fatalf("goroutine error: %v", err)
		}
	case <-done:
	}
}

func (suite *StreamTestSuite) TestStream_fillDataToReadBuffer() {
	conf := testConf()
	client, server := testClientServerConfig(conf)
	defer func() { _ = client.Close() }()
	defer func() { _ = server.Close() }()

	stream, _ := client.OpenStream()
	size := 8192
	slice := newBufferSlice(nil, make([]byte, size), 0, false)
	// Assign the return value of the first call to _ as it's not checked directly.
	_ = stream.fillDataToReadBuffer(bufferSliceWrapper{fallbackSlice: slice})
	assert.Equal(suite.T(), 1, len(stream.pendingData.unread))
	if err := stream.Close(); err != nil {
		suite.T().Logf("stream.Close error: %v", err)
	}
	// Keep the second call assigning to err as it is checked.
	err := stream.fillDataToReadBuffer(bufferSliceWrapper{fallbackSlice: slice})
	assert.Equal(suite.T(), 0, len(stream.pendingData.unread))
	assert.Equal(suite.T(), nil, err)

	//TODO: fillDataToReadBuffer 53.3%
}

// TestStream_SwapBufferForReuse is not yet implemented.
// TODO: Implement this test or remove if not needed.

/*
TODO: moveToWithoutLock and readMore tests are not yet implemented.
Remove or implement as needed.
*/

func TestStreamTestSuite(t *testing.T) {
	suite.Run(t, new(StreamTestSuite))
}
