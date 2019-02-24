//
// Copyright 2015 PingCAP, Inc.
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
//https://github.com/pingcap/tidb/blob/master/executor/adapter.go

package executor

import (
	"time"

	"github.com/pingcap/parser/ast"

	plannercore "fedb/planner/core"
	"fedb/sessionctx"
)

// ExecStmt implements the sqlexec.Statement interface, it builds a planner.Plan to an sqlexec.Statement.
type ExecStmt struct {
	// InfoSchema stores a reference to the schema information.
	//InfoSchema infoschema.InfoSchema
	// Plan stores a reference to the final physical plan.
	Plan plannercore.Plan
	// Expensive represents whether this query is an expensive one.
	//Expensive bool
	// Cacheable represents whether the physical plan can be cached.
	//Cacheable bool
	// Text represents the origin query text.
	Text string

	StmtNode ast.StmtNode

	Ctx sessionctx.Context
	// StartTime stands for the starting time when executing the statement.
	StartTime time.Time
	//isPreparedStmt bool
}
