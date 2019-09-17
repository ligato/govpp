//  Copyright (c) 2019 Cisco and/or its affiliates.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at:
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package statsclient

import (
	"fmt"
	"net"
	"sync/atomic"
	"syscall"
	"time"
	"unsafe"

	"github.com/ftrvxmtrx/fd"

	"git.fd.io/govpp.git/adapter"
)

var (
	MaxWaitInProgress    = time.Millisecond * 500
	CheckDelayInProgress = time.Microsecond * 100
)

const (
	MinVersion = 0
	MaxVersion = 1
)

func checkVersion(ver uint64) error {
	if ver < MinVersion {
		return fmt.Errorf("stat segment version is too old: %v (minimal version: %v)", ver, MinVersion)
	} else if ver > MaxVersion {
		return fmt.Errorf("stat segment version is not supported: %v (minimal version: %v)", ver, MaxVersion)
	}
	return nil
}

type statDirectoryType int32

func (t statDirectoryType) String() string {
	return adapter.StatType(t).String()
}

type statSegDirectoryEntry struct {
	directoryType statDirectoryType
	// unionData can represent:
	// - offset
	// - index
	// - value
	unionData    uint64
	offsetVector uint64
	name         [128]byte
}

type statSegment struct {
	sharedHeader []byte
	memorySize   int64

	// oldHeader defines version 0 for stat segment
	// and is used for VPP 19.04
	oldHeader bool
}

func (c *statSegment) connect(sockName string) error {
	addr := &net.UnixAddr{
		Net:  "unixpacket",
		Name: sockName,
	}

	Log.Debugf("connecting to: %v", addr)

	conn, err := net.DialUnix(addr.Net, nil, addr)
	if err != nil {
		Log.Warnf("connecting to socket %s failed: %s", addr, err)
		return err
	}
	defer func() {
		if err := conn.Close(); err != nil {
			Log.Warnf("closing socket failed: %v", err)
		}
	}()

	Log.Debugf("connected to socket")

	files, err := fd.Get(conn, 1, nil)
	if err != nil {
		return fmt.Errorf("getting file descriptor over socket failed: %v", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no files received over socket")
	}
	defer func() {
		for _, f := range files {
			if err := f.Close(); err != nil {
				Log.Warnf("closing file %s failed: %v", f.Name(), err)
			}
		}
	}()

	Log.Debugf("received %d files over socket", len(files))

	f := files[0]

	info, err := f.Stat()
	if err != nil {
		return err
	}
	size := info.Size()

	Log.Debugf("fd size=%v", size)

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		Log.Warnf("mapping shared memory failed: %v", err)
		return fmt.Errorf("mapping shared memory failed: %v", err)
	}

	c.sharedHeader = data
	c.memorySize = size

	Log.Debugf("successfuly mapped shared memory segment")

	header := c.readHeader()
	Log.Debugf("stat segment header: %+v", header)

	// older VPP (<=19.04) did not have version in stat segment header
	// we try to provide fallback support by skipping it in header
	if header.version > MaxVersion && header.inProgress > 1 && header.epoch == 0 {
		h := c.readHeaderOld()
		Log.Debugf("statsclient: falling back to old stat segment version (VPP <=19.04): %+v", h)
		c.oldHeader = true
	}

	return nil
}

func (c *statSegment) disconnect() error {
	if c.sharedHeader == nil {
		return nil
	}

	if err := syscall.Munmap(c.sharedHeader); err != nil {
		Log.Debugf("unmapping shared memory failed: %v", err)
		return fmt.Errorf("unmapping shared memory failed: %v", err)
	}
	c.sharedHeader = nil

	Log.Debugf("successfuly unmapped shared memory")
	return nil
}

func (c *statSegment) copyData(dirEntry *statSegDirectoryEntry) adapter.Stat {
	dirType := adapter.StatType(dirEntry.directoryType)

	switch dirType {
	case adapter.ScalarIndex:
		return adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorIndex:
		_, errOffset, _ := c.readOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])

		var errData adapter.Counter
		if c.oldHeader {
			// error were not vector (per-worker) in VPP 19.04
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(uint64(0))
			val := *(*adapter.Counter)(addOffset(offsetVector, offset))
			errData = val
		} else {
			vecLen := vectorLen(offsetVector)
			for i := uint64(0); i < vecLen; i++ {
				cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
				offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(addOffset(unsafe.Pointer(&c.sharedHeader[0]), offset))
				errData += val
			}
		}
		return adapter.ErrorStat(errData)

	case adapter.SimpleCounterVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		//simpleCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.Counter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			data[i] = make([]adapter.Counter, vecLen2)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(addOffset(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.SimpleCounterStat(data)

	case adapter.CombinedCounterVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		//combinedCounter := unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]) // offset
		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([][]adapter.CombinedCounter, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			data[i] = make([]adapter.CombinedCounter, vecLen2)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(addOffset(counterVec, offset))
				data[i][j] = val
			}
		}
		return adapter.CombinedCounterStat(data)

	case adapter.NameVector:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		//nameVector := // offset
		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := make([]adapter.Name, vecLen)
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				debugf("name vector out of range for %s (%v)", dirEntry.name, i)
				continue
			}
			nameVec := unsafe.Pointer(&c.sharedHeader[cb])
			vecLen2 := vectorLen(nameVec)

			nameStr := make([]byte, 0, vecLen2)
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(addOffset(nameVec, offset))
				if val > 0 {
					nameStr = append(nameStr, val)
				}
			}
			data[i] = adapter.Name(nameStr)
		}
		return adapter.NameStat(data)

	default:
		// TODO: monitor occurrences with metrics
		debugf("Unknown type %d for stat entry: %q", dirEntry.directoryType, dirEntry.name)
	}

	return nil
}

var ErrStatDataLenIncorrect = fmt.Errorf("stat data length incorrect")

//func (c *statSegment) updateStatEntry(dirEntry *statSegDirectoryEntry, stat *adapter.StatEntry) error {
func (c *statSegment) updateData(dirEntry *statSegDirectoryEntry, stat adapter.Stat) error {
	switch stat.(type) {
	case adapter.ScalarStat:
		stat = adapter.ScalarStat(dirEntry.unionData)

	case adapter.ErrorStat:
		_, errOffset, _ := c.readOffsets()
		offsetVector := unsafe.Pointer(&c.sharedHeader[errOffset])

		var errData adapter.Counter
		if c.oldHeader {
			// error were not vector (per-worker) in VPP 19.04
			offset := uintptr(dirEntry.unionData) * unsafe.Sizeof(uint64(0))
			val := *(*adapter.Counter)(addOffset(offsetVector, offset))
			errData = val
		} else {
			vecLen := vectorLen(offsetVector)
			for i := uint64(0); i < vecLen; i++ {
				cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
				offset := uintptr(cb) + uintptr(dirEntry.unionData)*unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(addOffset(unsafe.Pointer(&c.sharedHeader[0]), offset))
				errData += val
			}
		}
		stat = adapter.ErrorStat(errData)

	case adapter.SimpleCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := stat.(adapter.SimpleCounterStat)
		if uint64(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			simpData := data[i]
			if uint64(len(simpData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.Counter(0))
				val := *(*adapter.Counter)(addOffset(counterVec, offset))
				simpData[j] = val
			}
		}

	case adapter.CombinedCounterStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := stat.(adapter.CombinedCounterStat)
		if uint64(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			counterVec := unsafe.Pointer(&c.sharedHeader[uintptr(cb)])
			vecLen2 := vectorLen(counterVec)
			combData := data[i]
			if uint64(len(combData)) != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(adapter.CombinedCounter{})
				val := *(*adapter.CombinedCounter)(addOffset(counterVec, offset))
				combData[j] = val
			}
		}

	case adapter.NameStat:
		if dirEntry.unionData == 0 {
			debugf("offset invalid for %s", dirEntry.name)
			break
		} else if dirEntry.unionData >= uint64(len(c.sharedHeader)) {
			debugf("offset out of range for %s", dirEntry.name)
			break
		}

		vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirEntry.unionData]))
		offsetVector := addOffset(unsafe.Pointer(&c.sharedHeader[0]), uintptr(dirEntry.offsetVector))

		data := stat.(adapter.NameStat)
		if uint64(len(data)) != vecLen {
			return ErrStatDataLenIncorrect
		}
		for i := uint64(0); i < vecLen; i++ {
			cb := *(*uint64)(addOffset(offsetVector, uintptr(i)*unsafe.Sizeof(uint64(0))))
			if cb == 0 {
				continue
			}
			nameVec := unsafe.Pointer(&c.sharedHeader[cb])
			vecLen2 := vectorLen(nameVec)

			nameData := data[i]
			if uint64(len(nameData))+1 != vecLen2 {
				return ErrStatDataLenIncorrect
			}
			for j := uint64(0); j < vecLen2; j++ {
				offset := uintptr(j) * unsafe.Sizeof(byte(0))
				val := *(*byte)(addOffset(nameVec, offset))
				if val == 0 {
					break
				}
				nameData[j] = val
			}
		}

	default:
		/*if Debug {
			debugf("Unrecognized stat type %T for stat entry: %v", stat, dirEntry.name)
		}*/
		//return fmt.Errorf("unrecognized stat type %T", stat)
	}
	return nil
}

type sharedHeaderBase struct {
	epoch           int64
	inProgress      int64
	directoryOffset int64
	errorOffset     int64
	statsOffset     int64
}

type statSegSharedHeader struct {
	version uint64
	sharedHeaderBase
}

func (c *statSegment) readHeaderOld() (header statSegSharedHeader) {
	h := (*sharedHeaderBase)(unsafe.Pointer(&c.sharedHeader[0]))
	header.version = 0
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

func (c *statSegment) readHeader() (header statSegSharedHeader) {
	h := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	header.version = atomic.LoadUint64(&h.version)
	header.epoch = atomic.LoadInt64(&h.epoch)
	header.inProgress = atomic.LoadInt64(&h.inProgress)
	header.directoryOffset = atomic.LoadInt64(&h.directoryOffset)
	header.errorOffset = atomic.LoadInt64(&h.errorOffset)
	header.statsOffset = atomic.LoadInt64(&h.statsOffset)
	return
}

func (c *statSegment) readVersion() uint64 {
	if c.oldHeader {
		return 0
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	version := atomic.LoadUint64(&header.version)
	return version
}

func (c *statSegment) readEpoch() (int64, bool) {
	if c.oldHeader {
		h := c.readHeaderOld()
		return h.epoch, h.inProgress != 0
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	epoch := atomic.LoadInt64(&header.epoch)
	inprog := atomic.LoadInt64(&header.inProgress)
	return epoch, inprog != 0
}

func (c *statSegment) readOffsets() (dir, err, stat int64) {
	if c.oldHeader {
		h := c.readHeaderOld()
		return h.directoryOffset, h.errorOffset, h.statsOffset
	}
	header := (*statSegSharedHeader)(unsafe.Pointer(&c.sharedHeader[0]))
	dirOffset := atomic.LoadInt64(&header.directoryOffset)
	errOffset := atomic.LoadInt64(&header.errorOffset)
	statOffset := atomic.LoadInt64(&header.statsOffset)
	return dirOffset, errOffset, statOffset
}

type statSegAccess struct {
	epoch int64
}

func (c *statSegment) accessStart() statSegAccess {
	t := time.Now()
	epoch, inprog := c.readEpoch()
	for inprog {
		if time.Since(t) > MaxWaitInProgress {
			return statSegAccess{}
		} else {
			time.Sleep(CheckDelayInProgress)
		}
		epoch, inprog = c.readEpoch()
	}
	return statSegAccess{
		epoch: epoch,
	}
}

func (c *statSegment) accessEnd(acc *statSegAccess) bool {
	epoch, inprog := c.readEpoch()
	if acc.epoch != epoch || inprog {
		return false
	}
	return true
}

type vecHeader struct {
	length     uint64
	vectorData [0]uint8
}

func vectorLen(v unsafe.Pointer) uint64 {
	vec := *(*vecHeader)(unsafe.Pointer(uintptr(v) - unsafe.Sizeof(uintptr(0))))
	return vec.length
}

//go:nosplit
func addOffset(p unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + offset)
}
