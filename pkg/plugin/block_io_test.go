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
	"net"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlockReadFullAndBlockWriteFull(t *testing.T) {
	content := "hello,shmipc!"
	// Create a local Unix socket listener
	laddr, err := net.ResolveUnixAddr("unix", "/tmp/testBlockRWFull.sock")
	if err != nil {
		t.Fatalf("failed to resolve unix address: %v\n", err)
	}
	listener, err := net.ListenUnix("unix", laddr)
	if err != nil {
		t.Fatalf("failed to listen unix: %v\n", err)
		return
	}
	defer func() {
		listener.Close()                       //nolint:errcheck // test cleanup, error ignored intentionally
		os.Remove("/tmp/testBlockRWFull.sock") //nolint:errcheck // test cleanup, error ignored intentionally
	}()

	// Start a goroutine to accept a connection and write data
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			t.Errorf("failed to accept connection: %v\n", err)
		}
		defer func() {
			if err := conn.Close(); err != nil {
				t.Errorf("conn.Close failed: %v", err)
			}
		}()
		fd, err := getConnDupFd(conn)
		if err != nil {
			t.Errorf("failed to getConnDupFd: %v", err)
		}

		// Write data using blockWriteFull
		data := []byte(content)
		if err := blockWriteFull(int(fd.Fd()), data); err != nil {
			t.Errorf("failed to write data: %v\n", err)
		}
	}()

	// Dial the Unix socket and read data
	conn, err := net.DialUnix("unix", nil, laddr)
	if err != nil {
		t.Errorf("failed to dial unix: %v\n", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			// This Fatalf is okay as it's in the main test goroutine's defer
			t.Fatalf("conn.Close failed: %v", err)
		}
	}()
	fd, err := getConnDupFd(conn)
	if err != nil {
		t.Errorf("failed to getConnDupFd: %v", err)
	}

	// Read data using blockReadFull
	buf := make([]byte, 1024)
	if err := blockReadFull(int(fd.Fd()), buf[:len(content)]); err != nil {
		t.Errorf("failed to read data: %v\n", err)
	}
	// Check if the read data is correct
	assert.Equal(t, buf[:len(content)], []byte(content))

}
