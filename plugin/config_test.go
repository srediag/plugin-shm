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
	"time"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (s *ConfigTestSuite) TestVerifyConfig() {
	config := DefaultConfig()
	config.ShareMemoryBufferCap = 1
	err := VerifyConfig(config)
	s.Require().NotNil(err)
	config.ShareMemoryBufferCap = 1 << 20

	config.BufferSliceSizes = []*SizePercentPair{}
	err = VerifyConfig(config)
	s.Require().NotNil(err)
	config.BufferSliceSizes = []*SizePercentPair{
		{4096, 70},
		{16 << 10, 20},
		{64 << 10, 9},
	}
	err = VerifyConfig(config)
	s.Require().NotNil(err)

	config.BufferSliceSizes = []*SizePercentPair{
		{4096, 70},
		{16 << 10, 20},
		{64 << 10, 11},
	}
	err = VerifyConfig(config)
	s.Require().NotNil(err)

	config.BufferSliceSizes = []*SizePercentPair{
		{4096, 70},
		{16 << 10, 20},
		{defaultShareMemoryCap, 11},
	}
	err = VerifyConfig(config)
	s.Require().NotNil(err)

	config.BufferSliceSizes = []*SizePercentPair{
		{4096, 70},
		{16 << 10, 20},
		{64 << 10, 10},
	}
	err = VerifyConfig(config)
	s.Require().Nil(err)
}

func (s *ConfigTestSuite) TestCreateCSByWrongConfig() {
	conn1, conn2 := testConn()
	config := DefaultConfig()
	config.ShareMemoryBufferCap = 1
	c, err := newSession(config, conn1, true)
	s.Require().NotNil(err)
	s.Require().Nil(c)

	ok := make(chan struct{})
	go func() {
		sess, err := Server(conn2, config)
		s.Require().NotNil(err)
		s.Require().Nil(sess)
		close(ok)
	}()
	<-ok
}

func (s *ConfigTestSuite) TestCreateCSWithoutConfig() {
	conn1, conn2 := testConn()
	ok := make(chan struct{})
	go func() {
		sess, err := Server(conn2, nil)
		s.Require().Nil(err)
		s.Require().NotNil(sess)
		if err == nil {
			defer func() {
				if err := sess.Close(); err != nil {
					s.T().Fatalf("sess.Close failed: %v", err)
				}
			}()
		}
		close(ok)
	}()

	c, err := newSession(nil, conn1, true)
	s.Require().Nil(err)
	if err == nil {
		defer func() {
			if err := c.Close(); err != nil {
				s.T().Fatalf("c.Close failed: %v", err)
			}
		}()
	}
	s.Require().NotNil(c)
	time.Sleep(time.Second)
	<-ok
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
