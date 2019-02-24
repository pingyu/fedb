//
// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.
//
//https://github.com/pingcap/tidb/blob/master/store/tikv/oracle/oracle.go

package oracle

import (
	"context"
	"time"
)

// Oracle is the interface that provides strictly ascending timestamps.
type Oracle interface {
	GetTimestamp(ctx context.Context) (uint64, error)
	GetTimestampAsync(ctx context.Context) Future
	//IsExpired(lockTimestamp uint64, TTL uint64) bool
	Close()
}

// Future is a future which promises to return a timestamp.
type Future interface {
	Wait() (uint64, error)
}

const physicalShiftBits = 18

// ComposeTS creates a ts from physical and logical parts.
func ComposeTS(physical, logical int64) uint64 {
	return uint64((physical << physicalShiftBits) + logical)
}

// GetPhysical returns physical from an instant time with millisecond precision.
func GetPhysical(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}
