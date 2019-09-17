// Copyright (c) 2019 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package statsclient is pure Go implementation of VPP stats API client.
package statsclient

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"unsafe"

	logger "github.com/sirupsen/logrus"

	"git.fd.io/govpp.git/adapter"
)

const (
	// DefaultSocketName is default VPP stats socket file path.
	DefaultSocketName = adapter.DefaultStatsSocket
)

const socketMissing = `
------------------------------------------------------------
 VPP stats socket file %s is missing!

  - is VPP running with stats segment enabled?
  - is the correct socket name configured?

 To enable it add following section to your VPP config:
   statseg {
     default
   }
------------------------------------------------------------
`

var (
	// Debug is global variable that determines debug mode
	Debug = os.Getenv("DEBUG_GOVPP_STATS") != ""

	// Log is global logger
	Log = logger.New()
)

// init initializes global logger, which logs debug level messages to stdout.
func init() {
	Log.Out = os.Stdout
	if Debug {
		Log.Level = logger.DebugLevel
		Log.Debug("govpp/statsclient: enabled debug mode")
	}
}

func debugf(f string, a ...interface{}) {
	if Debug {
		Log.Debugf(f, a...)
	}
}

// implements StatsAPI
var _ adapter.StatsAPI = (*StatsClient)(nil)

// StatsClient is the pure Go implementation for VPP stats API.
type StatsClient struct {
	sockAddr string

	currentEpoch int64
	statSegment
}

// NewStatsClient returns new VPP stats API client.
func NewStatsClient(sockAddr string) *StatsClient {
	if sockAddr == "" {
		sockAddr = DefaultSocketName
	}
	return &StatsClient{
		sockAddr: sockAddr,
	}
}

func (c *StatsClient) Connect() error {
	// check if socket exists
	if _, err := os.Stat(c.sockAddr); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, socketMissing, c.sockAddr)
		return fmt.Errorf("stats socket file %s does not exist", c.sockAddr)
	} else if err != nil {
		return fmt.Errorf("stats socket error: %v", err)
	}

	if err := c.statSegment.connect(c.sockAddr); err != nil {
		return err
	}

	ver := c.readVersion()
	Log.Debugf("stat segment version: %v", ver)

	if err := checkVersion(ver); err != nil {
		return err
	}
	return nil
}

func (c *StatsClient) Disconnect() error {
	if err := c.statSegment.disconnect(); err != nil {
		return err
	}
	return nil
}

func (c *StatsClient) ListStats(patterns ...string) (names []string, err error) {
	dir, err := c.listIndexesRegex(patterns...)
	if err != nil {
		return nil, err
	}
	for _, d := range dir {
		name, err := c.GetStatName(d)
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, nil
}

func (c *StatsClient) DumpStats(patterns ...string) (entries []adapter.StatEntry, err error) {
	dir, err := c.listIndexesRegex(patterns...)
	if err != nil {
		return nil, err
	}
	if entries, err = c.DumpEntries(dir); err != nil {
		return nil, err
	}
	return entries, nil
}

func (c *StatsClient) listIndexesFn(matchFn func(name []byte) bool) (indexes []uint32, err error) {
	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		var name []byte
		for i := 0; i < len(dirEntry.name); i++ {
			if dirEntry.name[i] == 0x00 {
				name = dirEntry.name[:i]
				break
			}
		}
		if len(name) == 0 {
			Log.Debugf("invalid entry name found for index %d: %q", i, dirEntry.name[:])
			continue
		}

		if matchFn == nil || matchFn(name) {
			indexes = append(indexes, uint32(i))
		}
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatDirBusy
	}
	c.currentEpoch = sa.epoch

	return indexes, nil
}

func (c *StatsClient) listIndexesRegex(patterns ...string) (indexes []uint32, err error) {
	var regexes = make([]*regexp.Regexp, len(patterns))
	for i, pattern := range patterns {
		r, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("compiling regexp failed: %v", err)
		}
		regexes[i] = r
	}
	nameMatches := func(name []byte) bool {
		for _, r := range regexes {
			if r.Match(name) {
				return true
			}
		}
		return false
	}
	return c.listIndexesFn(nameMatches)
}

func (c *StatsClient) GetStatName(index uint32) (string, error) {
	sa := c.accessStart()
	if sa.epoch == 0 {
		return "", fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()
	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))

	i := uint64(index)
	if i >= vecLen {
		return "", fmt.Errorf("dir index %d out of range (%d)", i, vecLen)
	}

	offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
	dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

	nul := bytes.IndexByte(dirEntry.name[:], '\x00')
	if nul < 0 {
		Log.Debugf("no zero byte found for: %q", dirEntry.name[:])
		return "", fmt.Errorf("invalid name for stat dir")
	}
	name := dirEntry.name[:nul]
	if len(name) == 0 {
		return "", fmt.Errorf("invalid name for stat dir")
	}

	if !c.accessEnd(&sa) {
		return "", adapter.ErrStatDirBusy
	}

	return string(name), nil
}

func (c *StatsClient) PrepareDir(prefixes ...string) (*adapter.StatDir, error) {
	var matches func(name []byte) bool
	if len(prefixes) > 0 {
		matches = func(name []byte) bool {
			for _, prefix := range prefixes {
				if bytes.HasPrefix(name, []byte(prefix)) {
					return true
				}
			}
			return false
		}
	}

	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, fmt.Errorf("access failed")
	}

	dir := &adapter.StatDir{
		Epoch: sa.epoch,
	}

	dirOffset, _, _ := c.readOffsets()
	vecLen := vectorLen(unsafe.Pointer(&c.sharedHeader[dirOffset]))

	for i := uint64(0); i < vecLen; i++ {
		offset := uintptr(i) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		if matches == nil || matches(dirEntry.name[:]) {
			dir.Indexes = append(dir.Indexes, uint32(i))
		}
	}

	dir.Entries = make([]adapter.StatEntry, len(dir.Indexes))

	for i, index := range dir.Indexes {
		offset := uintptr(index) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		nul := bytes.IndexByte(dirEntry.name[:], 0x00)
		if nul < 0 {
			Log.Debugf("no zero byte found for: %q", dirEntry.name[:])
			continue
		}
		name := dirEntry.name[:nul]
		if len(name) == 0 {
			continue
		}

		dir.Entries[i] = adapter.StatEntry{
			Name: append([]byte(nil), name...),
			Type: adapter.StatType(dirEntry.directoryType),
			Data: c.copyData(dirEntry),
		}
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatDirBusy
	}

	return dir, nil
}

func (c *StatsClient) UpdateDir(dir *adapter.StatDir) (err error) {
	epoch, _ := c.readEpoch()
	if dir.Epoch != epoch {
		return adapter.ErrStatDataStale
	}

	sa := c.accessStart()
	if sa.epoch == 0 {
		return fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()

	for i, index := range dir.Indexes {
		offset := uintptr(index) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		var name []byte
		for n := 0; i < len(dirEntry.name); n++ {
			if dirEntry.name[n] == 0x00 {
				name = dirEntry.name[:n]
				break
			}
		}
		if len(name) == 0 {
			continue
		}

		entry := dir.Entries[i]
		if !bytes.Equal(name, entry.Name) {
			continue
		}
		if adapter.StatType(dirEntry.directoryType) != entry.Type {
			continue
		}
		if entry.Data == nil {
			continue
		}
		if err := c.updateData(dirEntry, entry.Data); err != nil {
			return fmt.Errorf("updating stat data for entry %s failed: %v", name, err)
		}
	}

	if !c.accessEnd(&sa) {
		return adapter.ErrStatDumpBusy
	}
	return nil
}

func (c *StatsClient) DumpEntries(indexes []uint32) (entries []adapter.StatEntry, err error) {
	epoch, _ := c.readEpoch()
	if c.currentEpoch > 0 && c.currentEpoch != epoch {
		return nil, fmt.Errorf("stat data stale")
	}

	sa := c.accessStart()
	if sa.epoch == 0 {
		return nil, fmt.Errorf("access failed")
	}

	dirOffset, _, _ := c.readOffsets()

	entries = make([]adapter.StatEntry, 0, len(indexes))

	for _, index := range indexes {
		offset := uintptr(index) * unsafe.Sizeof(statSegDirectoryEntry{})
		dirEntry := (*statSegDirectoryEntry)(addOffset(unsafe.Pointer(&c.sharedHeader[dirOffset]), offset))

		var name []byte
		for i := 0; i < len(dirEntry.name); i++ {
			if dirEntry.name[i] == 0x00 {
				name = dirEntry.name[:i]
				break
			}
		}
		if len(name) == 0 {
			Log.Debugf("invalid entry name found for index %d: %q", index, dirEntry.name[:])
		}

		entry := adapter.StatEntry{
			Name: append([]byte(nil), name...),
			Type: adapter.StatType(dirEntry.directoryType),
			Data: c.copyData(dirEntry),
		}

		entries = append(entries, entry)
	}

	if !c.accessEnd(&sa) {
		return nil, adapter.ErrStatDumpBusy
	}

	return entries, nil
}
