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
//https://github.com/pingcap/tidb/blob/master/planner/core/plan.go

package core

import (
	"fedb/sessionctx"
)

// Plan is the description of an execution flow.
// It is created from ast.Node first, then optimized by the optimizer,
// finally used by the executor to create a Cursor which executes the statement.
type Plan interface {
	// Get the schema.
	//Schema() *expression.Schema
	// Get the ID.
	ID() int
	// Get the ID in explain statement
	ExplainID() string
	// replaceExprColumns replace all the column reference in the plan's expression node.
	//replaceExprColumns(replace map[string]*expression.Column)

	context() sessionctx.Context

	// property.StatsInfo will return the property.StatsInfo for this plan.
	//statsInfo() *property.StatsInfo
}
