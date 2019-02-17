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
// some code copied from Copyright 2016 PingCAP, Inc.
// https://github.com/pingcap/tidb/blob/source-code/server/driver_tidb.go
//

package server

import (
	"crypto/tls"
	"fedb/util/sqlexec"

	"github.com/pingcap/errors"
	goctx "golang.org/x/net/context"

	"fedb/session"
)

// FeDBDriver implements IDriver.
type FeDBDriver struct {
	//store
}

// NewFeDBDriver creates a new FeDBDriver.
func NewFeDBDriver() *FeDBDriver {
	driver := &FeDBDriver{}
	return driver
}

// OpenCtx creates context
func (drv *FeDBDriver) OpenCtx(connID uint64, capability uint32, collation uint8, dbname string, tlsState *tls.ConnectionState) (QueryCtx, error) {
	//TODO: ignore collation, tlsState

	session, err := session.CreateSession()
	if err != nil {
		return nil, errors.Trace(err)
	}

	session.SetConnectionID(connID).SetClientCapability(capability)
	err = session.SetCollation(int(collation))
	if err != nil {
		return nil, errors.Trace(err)
	}

	ctx := &FeDBContext{
		session:   session,
		currentDB: dbname,
	}
	return ctx, nil
}

// FeDBContext implements QueryCtx.
type FeDBContext struct {
	session   session.Session
	currentDB string
}

type fedbResultSet struct {
	recordSet sqlexec.RecordSet
	//columns   []*ColumnInfo
	//rows      []chunk.Row
	//closed    bool
}

// Execute executes SQL query
func (ctx *FeDBContext) Execute(goCtx goctx.Context, sql string) (rs []ResultSet, err error) {
	rsList, err := ctx.session.Execute(goCtx, sql)
	if err != nil {
		return nil, err
	}
	if len(rsList) == 0 {
		return nil, nil
	}

	rs = make([]ResultSet, len(rsList))
	for i := 0; i < len(rsList); i++ {
		rs[i] = &fedbResultSet{
			recordSet: rsList[i],
		}
	}
	return rs, nil
}

// Close closes context
func (ctx *FeDBContext) Close() error {
	ctx.session.Close()
	return nil
}
