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
	"os"
	"path/filepath"
	"syscall"
	"unsafe"

	queuepkg "github.com/Workiva/go-datastructures/queue"
)

const (
	queueHeaderLength = 24
)

type queueManager struct {
	path        string
	sendQueue   *queue
	recvQueue   *queue
	mem         []byte
	mmapMapType MemMapType
	memFd       int
}

// default cap is 16384, which mean that 16384 * 8 = 128 KB memory.
type queue struct {
	q *queuepkg.Queue
}

type queueElement struct {
	seqID          uint32
	offsetInShmBuf uint32
	status         uint32
}

func createQueueManagerWithMemFd(queuePathName string, queueCap uint32) (*queueManager, error) {
	memFd, err := MemfdCreate(queuePathName, 0)
	if err != nil {
		return nil, err
	}

	memSize := countQueueMemSize(queueCap) * queueCount
	if err := syscall.Ftruncate(memFd, int64(memSize)); err != nil {
		return nil, fmt.Errorf("createQueueManagerWithMemFd truncate share memory failed,%w", err)
	}

	mem, err := syscall.Mmap(memFd, 0, memSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(mem); i++ {
		mem[i] = 0
	}

	return &queueManager{
		sendQueue:   createQueueFromBytes(mem[:memSize/2], queueCap),
		recvQueue:   createQueueFromBytes(mem[memSize/2:], queueCap),
		mem:         mem,
		path:        queuePathName,
		mmapMapType: MemMapTypeMemFd,
		memFd:       memFd,
	}, nil
}

func createQueueManager(shmPath string, queueCap uint32) (*queueManager, error) {
	//ignore mkdir error
	_ = os.MkdirAll(filepath.Dir(shmPath), os.ModePerm)
	if pathExists(shmPath) {
		return nil, fmt.Errorf("queue was existed,path %s", shmPath)
	}
	memSize := countQueueMemSize(queueCap) * queueCount
	if !canCreateOnDevShm(uint64(memSize), shmPath) {
		return nil, fmt.Errorf("err:%s path:%s, size:%d", ErrShareMemoryHadNotLeftSpace.Error(), shmPath, memSize)
	}
	f, err := os.OpenFile(shmPath, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			internalLogger.warnf("file close error: %v", cerr)
		}
	}()

	if err := f.Truncate(int64(memSize)); err != nil {
		return nil, fmt.Errorf("truncate share memory failed,%s", err.Error())
	}
	mem, err := syscall.Mmap(int(f.Fd()), 0, memSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(mem); i++ {
		mem[i] = 0
	}
	return &queueManager{
		sendQueue: createQueueFromBytes(mem[:memSize/2], queueCap),
		recvQueue: createQueueFromBytes(mem[memSize/2:], queueCap),
		mem:       mem,
		path:      shmPath,
	}, nil
}

func mappingQueueManagerMemfd(queuePathName string, memFd int) (*queueManager, error) {
	var fileInfo syscall.Stat_t
	if err := syscall.Fstat(memFd, &fileInfo); err != nil {
		return nil, err
	}

	mappingSize := int(fileInfo.Size)
	//a queueManager have two queue, a queue's head and tail should align to 8 byte boundary
	if isArmArch() && mappingSize%16 != 0 {
		return nil, fmt.Errorf("the memory size of queue should be a multiple of 16")
	}
	mem, err := syscall.Mmap(memFd, 0, mappingSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return &queueManager{
		sendQueue:   mappingQueueFromBytes(mem[mappingSize/2:]),
		recvQueue:   mappingQueueFromBytes(mem[:mappingSize/2]),
		mem:         mem,
		path:        queuePathName,
		memFd:       memFd,
		mmapMapType: MemMapTypeMemFd,
	}, nil
}

func mappingQueueManager(shmPath string) (*queueManager, error) {
	f, err := os.OpenFile(shmPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		return nil, err
	}
	defer func() {
		cerr := f.Close()
		if cerr != nil {
			internalLogger.warnf("file close error: %v", cerr)
		}
	}()
	fileInfo, err := f.Stat()
	if err != nil {
		return nil, err
	}
	mappingSize := int(fileInfo.Size())

	//a queueManager have two queue, a queue's head and tail should align to 8 byte boundary
	if isArmArch() && mappingSize%16 != 0 {
		return nil, fmt.Errorf("the memory size of queue should be a multiple of 16")
	}
	mem, err := syscall.Mmap(int(f.Fd()), 0, mappingSize, syscall.PROT_READ|syscall.PROT_WRITE, syscall.MAP_SHARED)
	if err != nil {
		return nil, err
	}
	return &queueManager{
		sendQueue: mappingQueueFromBytes(mem[mappingSize/2:]),
		recvQueue: mappingQueueFromBytes(mem[:mappingSize/2]),
		mem:       mem,
		path:      shmPath,
	}, nil
}

func countQueueMemSize(queueCap uint32) int {
	return queueHeaderLength + queueElementLen*int(queueCap)
}

func createQueueFromBytes(_ []byte, cap uint32) *queue {
	return &queue{q: queuepkg.New(int64(cap))}
}

func mappingQueueFromBytes(data []byte) *queue {
	cap := *(*uint32)(unsafe.Pointer(&data[0]))
	return &queue{q: queuepkg.New(int64(cap))}
}

func (q *queue) pop() (e queueElement, err error) {
	items, err := q.q.Get(1)
	if err != nil || len(items) == 0 {
		return queueElement{}, err
	}
	qe, ok := items[0].(queueElement)
	if !ok {
		return queueElement{}, fmt.Errorf("invalid queue element type")
	}
	return qe, nil
}

func (q *queue) put(e queueElement) error {
	return q.q.Put(e)
}

func (q *queueManager) unmap() {
	if err := syscall.Munmap(q.mem); err != nil {
		internalLogger.warnf("queueManager unmap error:" + err.Error())
	}
	if q.mmapMapType == MemMapTypeDevShmFile {
		if err := os.Remove(q.path); err != nil {
			internalLogger.warnf("queueManager remove file:%s failed, error=%s", q.path, err.Error())
		} else {
			internalLogger.infof("queueManager remove file:%s", q.path)
		}
	} else {
		if err := syscall.Close(q.memFd); err != nil {
			internalLogger.warnf("queueManager close queue fd:%d, error:%s", q.memFd, err.Error())
		} else {
			internalLogger.infof("queueManager close queue fd:%d", q.memFd)
		}
	}
}
