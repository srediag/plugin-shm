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
	"math"
	"net/http"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/bytebufferpool"
)

func TestPathExists(t *testing.T) {
	path := "test_path_exists"
	f, err := os.OpenFile(path, os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()
	assert.Equal(t, true, pathExists(path))
	_ = os.Remove(path)
}

func TestCanCreateOnDevShm(t *testing.T) {
	switch runtime.GOOS {
	case "linux":
		//just on /dev/shm, other always return true
		assert.Equal(t, true, canCreateOnDevShm(math.MaxUint64, "sdffafds"))
		stat, err := disk.Usage("/dev/shm")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, true, canCreateOnDevShm(stat.Free, "/dev/shm/xxx"))
		assert.Equal(t, false, canCreateOnDevShm(stat.Free+1, "/dev/shm/yyy"))
	case "darwin":
		//always return true
		assert.Equal(t, true, canCreateOnDevShm(33333, "sdffafds"))
	}
}

func TestSafeRemoveUdsFile(t *testing.T) {
	path := "test_path_remove"
	f, err := os.OpenFile(path, os.O_CREATE, os.ModePerm)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.Close()

	assert.Equal(t, true, safeRemoveUdsFile(path))
	assert.Equal(t, false, safeRemoveUdsFile("not_existing_file"))
}

var (
	bufferAllocations = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "buffer_allocations_total",
		Help: "Total number of buffer allocations.",
	})
)

func init() {
	prometheus.MustRegister(bufferAllocations)
}

func TestByteBufferPoolIntegration(t *testing.T) {
	buf := bytebufferpool.Get()
	defer bytebufferpool.Put(buf)
	_, err := buf.WriteString("hello world")
	if err != nil {
		t.Fatalf("bytebufferpool: WriteString returned error: %v", err)
	}
	if buf.String() != "hello world" {
		t.Errorf("bytebufferpool: expected 'hello world', got '%s'", buf.String())
	}
}

func TestBackoffIntegration(t *testing.T) {
	attempts := 0
	op := func() error {
		attempts++
		if attempts < 3 {
			return assert.AnError
		}
		return nil
	}
	err := backoff.Retry(op, backoff.WithMaxRetries(backoff.NewConstantBackOff(10*time.Millisecond), 5))
	if err != nil {
		t.Errorf("backoff: expected success after retries, got error: %v", err)
	}
	if attempts != 3 {
		t.Errorf("backoff: expected 3 attempts, got %d", attempts)
	}
}

func TestPrometheusCounter(t *testing.T) {
	c := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_counter",
		Help: "A test counter.",
	})
	c.Inc()
	c.Add(2)
	if v := prometheusToFloat64(c); v != 3 {
		t.Errorf("prometheus: expected counter value 3, got %v", v)
	}
}

// prometheusToFloat64 extrai o valor de um Counter para teste
func prometheusToFloat64(c prometheus.Counter) float64 {
	m := &dto.Metric{}
	_ = c.Write(m)
	return m.GetCounter().GetValue()
}

func TestHealthcheckHandler(t *testing.T) {
	health := healthcheck.NewHandler()
	health.AddLivenessCheck("always-ok", func() error { return nil })
	// Simula chamada HTTP
	req, _ := http.NewRequest("GET", "/live", nil)
	rw := &testResponseWriter{}
	health.ServeHTTP(rw, req)
	if rw.status != 200 {
		t.Errorf("healthcheck: expected status 200, got %d", rw.status)
	}
}

type testResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (w *testResponseWriter) Header() http.Header {
	if w.headers == nil {
		w.headers = make(http.Header)
	}
	return w.headers
}

func (w *testResponseWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *testResponseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
}
