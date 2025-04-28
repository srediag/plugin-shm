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
	"testing"

	"github.com/stretchr/testify/suite"
)

type DebugTestSuite struct {
	suite.Suite
}

func (s *DebugTestSuite) TestLogColor() {
	SetLogLevel(levelTrace)

	internalLogger.tracef("this is tracef %s", "hello world")
	internalLogger.tracef("trace message")

	internalLogger.infof("this is infof %s", "hello world")
	internalLogger.info("this is info")

	internalLogger.debugf("this is debugf %s", "hello world")
	internalLogger.debugf("debug message")

	internalLogger.warnf("this is warnf %s", "hello world")
	internalLogger.warnf("warn message")

	internalLogger.errorf("this is errorf %s", "hello world")
	internalLogger.error("this is error")
}

func TestDebugTestSuite(t *testing.T) {
	suite.Run(t, new(DebugTestSuite))
}
