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

package shmipc_async_client

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/srediag/plugin-shm/examples/best_practice/idl"
	"github.com/srediag/plugin-shm/pkg/plugin"
)

var (
	count uint64
	_     plugin.StreamCallbacks = &streamCbImpl{}
)

func init() {
	go func() {
		lastCount := count
		for range time.Tick(time.Second) {
			curCount := atomic.LoadUint64(&count)
			fmt.Println("shmipc_async_client qps:", curCount-lastCount)
			lastCount = curCount
		}
	}()
	runtime.GOMAXPROCS(1)

	go func() {
		http.ListenAndServe(":20001", nil) //nolint:errcheck
	}()
}

type streamCbImpl struct {
	req    idl.Request
	resp   idl.Response
	stream *plugin.Stream
	key    []byte
	loop   uint64
	n      uint64
}

func (s *streamCbImpl) OnData(reader plugin.BufferReader) {
	//wait and read response
	s.resp.Reset()
	if err := s.resp.ReadFromShm(reader); err != nil {
		fmt.Println("write request to share memory failed, err=" + err.Error())
		return
	}
	s.stream.ReleaseReadAndReuse()

	{
		//handle response...
		atomic.AddUint64(&count, 1)
	}
	s.send()
}

func (s *streamCbImpl) send() {
	s.n++
	if s.n >= s.loop {
		return
	}
	now := time.Now()
	//serialize request
	s.req.Reset()
	s.req.ID = uint64(now.UnixNano())
	s.req.Name = "xxx"
	s.req.Key = s.key
	if err := s.req.WriteToShm(s.stream.BufferWriter()); err != nil {
		fmt.Println("write request to share memory failed, err=" + err.Error())
		return
	}
	if err := s.stream.Flush(false); err != nil {
		fmt.Println(" flush request to peer failed, err=" + err.Error())
		return
	}
}

func (s *streamCbImpl) OnLocalClose() {
	//fmt.Println("stream OnLocalClose")
}

func (s *streamCbImpl) OnRemoteClose() {
	//fmt.Println("stream OnRemoteClose")

}
