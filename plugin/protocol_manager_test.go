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
	_ "net/http/pprof"
	"testing"

	"github.com/stretchr/testify/suite"
)

type ProtocolManagerTestSuite struct {
	suite.Suite
}

func (s *ProtocolManagerTestSuite) TestProtocolCompatibilityForNetUnixConn() {
	s.testProtocolCompatibility(MemMapTypeMemFd)
}

func (s *ProtocolManagerTestSuite) testProtocolCompatibility(memType MemMapType) {
	fmt.Println("----------bengin test protocolAdaptor MemMapType ----------", memType)
	clientConn, serverConn := testUdsConn()
	conf := testConf()
	conf.MemMapType = memType
	go func() {
		sconf := testConf()
		server, err := Server(serverConn, sconf)
		s.Require().True(err == nil, err)
		if err == nil {
			cerr := server.Close()
			if cerr != nil {
				s.T().Errorf("server.Close error: %v", cerr)
			}
		}
	}()

	client, err := newSession(conf, clientConn, true)
	s.Require().True(err == nil, err)
	if err == nil {
		cerr := client.Close()
		if cerr != nil {
			s.T().Errorf("client.Close error: %v", cerr)
		}
	}

	fmt.Println("----------end test protocolAdaptor client V2 to server V2 ----------")
}

func TestProtocolManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ProtocolManagerTestSuite))
}
