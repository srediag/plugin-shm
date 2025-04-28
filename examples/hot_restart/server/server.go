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

package server

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
)

var (
	count    uint64
	errCount uint64

	sendStr = "hello world"

	udsPath   = ""
	adminPath = ""

	ENV_IS_HOT_RESTART_KEY    = "IS_HOT_RESTART"
	ENV_HOT_RESTART_EPOCH_KEY = "HOT_RESTART_EPOCH"
)

func Init() {
	go func() {
		lastCount := count
		for range time.Tick(time.Second) {
			curCount := atomic.LoadUint64(&count)
			fmt.Println("qps:", curCount-lastCount, " errcount ", atomic.LoadUint64(&errCount), " count ", atomic.LoadUint64(&count))
			lastCount = curCount
		}
	}()
	runtime.GOMAXPROCS(4)

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	udsPath = filepath.Join(dir, "../ipc_test.sock")
	adminPath = filepath.Join(dir, "../admin.sock")
	fmt.Printf("shmipc udsPath %s adminPath %s\n", udsPath, adminPath)

	debugPort := 20000
	if os.Getenv("DEBUG_PORT") != "" {
		debugPort, err = strconv.Atoi(os.Getenv("DEBUG_PORT"))
		if err != nil {
			panic(err)
		}
	}

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", debugPort), nil) //nolint:errcheck
	}()
}
