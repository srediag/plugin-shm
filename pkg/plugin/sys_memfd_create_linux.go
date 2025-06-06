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
	"golang.org/x/sys/unix"
)

// linux 3.17+ provided
func MemfdCreate(name string, flags int) (fd int, err error) {
	memFd, err := unix.MemfdCreate(memfdCreateName+name, 0)
	if err != nil {
		return 0, err
	}

	return memFd, nil
}
