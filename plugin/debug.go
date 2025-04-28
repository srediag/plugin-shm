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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

type logger struct {
	name      string
	out       io.Writer
	callDepth int
}

var (
	internalLogger = &logger{"", os.Stdout, 3}
	protocolLogger = &logger{"protocol trace", os.Stdout, 4}
	level          int
	debugMode      = false

	magenta = string([]byte{27, 91, 57, 53, 109}) // Trace
	green   = string([]byte{27, 91, 57, 50, 109}) // Debug
	blue    = string([]byte{27, 91, 57, 52, 109}) // Info
	yellow  = string([]byte{27, 91, 57, 51, 109}) // Warn
	red     = string([]byte{27, 91, 57, 49, 109}) // Error
	reset   = string([]byte{27, 91, 48, 109})

	colors = []string{
		magenta,
		green,
		blue,
		yellow,
		red,
	}

	levelName = []string{
		"Trace",
		"Debug",
		"Info",
		"Warn",
		"Error",
	}
)

const (
	levelTrace = iota
	levelDebug
	levelInfo
	levelWarn
	levelError
	levelNoPrint
)

func init() {
	level = levelWarn
	if os.Getenv("SHMIPC_LOG_LEVEL") != "" {
		if n, err := strconv.Atoi(os.Getenv("SHMIPC_LOG_LEVEL")); err == nil {
			if n <= levelNoPrint {
				level = n
			}
		}
	}

	if os.Getenv("SHMIPC_DEBUG_MODE") != "" {
		debugMode = true
	}
}

// SetLogLevel used to change the internal logger's level and the default level is Warning.
// The process env `SHMIPC_LOG_LEVEL` also could set log level
func SetLogLevel(l int) {
	if l <= levelNoPrint {
		level = l
	}
}

func newLogger(name string, out io.Writer) *logger {
	if out == nil {
		out = os.Stdout
	}
	return &logger{
		name:      name,
		out:       out,
		callDepth: 3,
	}
}

func (l *logger) errorf(format string, a ...interface{}) {
	if level > levelError {
		return
	}
	if _, err := fmt.Fprintf(l.out, l.prefix(levelError)+format+reset+"\n", a...); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger errorf failed: %v\n", err)
	}
}

func (l *logger) error(v interface{}) {
	if level > levelError {
		return
	}
	if _, err := fmt.Fprintln(l.out, l.prefix(levelError), v, reset); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger error failed: %v\n", err)
	}
}

func (l *logger) warnf(format string, a ...interface{}) {
	if level > levelWarn {
		return
	}
	if _, err := fmt.Fprintf(l.out, l.prefix(levelWarn)+format+reset+"\n", a...); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger warnf failed: %v\n", err)
	}
}

func (l *logger) infof(format string, a ...interface{}) {
	if level > levelInfo {
		return
	}
	if _, err := fmt.Fprintf(l.out, l.prefix(levelInfo)+format+reset+"\n", a...); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger infof failed: %v\n", err)
	}
}

func (l *logger) info(v interface{}) {
	if level > levelInfo {
		return
	}
	if _, err := fmt.Fprintln(l.out, l.prefix(levelInfo), v, reset); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger info failed: %v\n", err)
	}
}

func (l *logger) debugf(format string, a ...interface{}) {
	if level > levelDebug {
		return
	}
	if _, err := fmt.Fprintf(l.out, l.prefix(levelDebug)+format+reset+"\n", a...); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger debugf failed: %v\n", err)
	}
}

func (l *logger) tracef(format string, a ...interface{}) {
	if level > levelTrace {
		return
	}
	if _, err := fmt.Fprintf(l.out, l.prefix(levelTrace)+format+reset+"\n", a...); err != nil {
		// Optionally log or handle the error
		fmt.Fprintf(os.Stderr, "logger tracef failed: %v\n", err)
	}
}

func (l *logger) prefix(level int) string {
	var buffer [64]byte
	buf := bytes.NewBuffer(buffer[:0])
	_, _ = buf.WriteString(colors[level])
	_, _ = buf.WriteString(levelName[level])
	_ = buf.WriteByte(' ')
	_, _ = buf.WriteString(time.Now().Format("2006-01-02 15:04:05.999999"))
	_ = buf.WriteByte(' ')
	_, _ = buf.WriteString(l.location())
	_ = buf.WriteByte(' ')
	_, _ = buf.WriteString(l.name)
	_ = buf.WriteByte(' ')
	return buf.String()
}

func (l *logger) location() string {
	_, file, line, ok := runtime.Caller(l.callDepth)
	if !ok {
		file = "???"
		line = 0
	}
	file = filepath.Base(file)
	return file + ":" + strconv.Itoa(line)
}

// DebugQueueDetail print IO-Queue's status which was mmap in the `path`
func DebugQueueDetail(path string) {
	mem, err := os.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	sendQueue := mappingQueueFromBytes(mem[len(mem)/2:])
	recvQueue := mappingQueueFromBytes(mem[:len(mem)/2])
	printFunc := func(name string, q *queue) {
		fmt.Printf("path:%s name:%s, cap:%d head:%d tail:%d size:%d flag:%d\n",
			name, path, q.cap, *q.head, *q.tail, q.size(), *q.workingFlag)
	}
	printFunc("sendQueue", sendQueue)
	printFunc("recvQueue", recvQueue)
}
