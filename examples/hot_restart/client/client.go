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

package client

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	count    uint64
	errCount uint64
)

func init() {
	go func() {
		lastCount := atomic.LoadUint64(&count)
		for range time.Tick(time.Second) {
			curCount := atomic.LoadUint64(&count)
			err := atomic.LoadUint64(&errCount)
			fmt.Println("qps:", curCount-lastCount, "  errCount = ", err, " count ", atomic.LoadUint64(&count))
			lastCount = curCount
		}
	}()
	runtime.GOMAXPROCS(4)

	go func() {
		http.ListenAndServe(":20001", nil) //nolint:errcheck
	}()
}
