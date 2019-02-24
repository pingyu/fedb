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
//https://github.com/pingcap/tidb/blob/master/executor/compiler.go

package executor

import (
	"context"

	"github.com/opentracing/opentracing-go"

	"fedb/sessionctx"
)

// Compiler compiles an ast.StmtNode to a physical plan.
type Compiler struct {
	Ctx sessionctx.Context
}

// Compile an ast.StmtNode to a physical plan.
func (c *Complier) Compile(ctx context.Context, stmtNode ast.StmtNode) (*ExecStmt, error) {
	if span := opentracing.SpanFromContext(ctx); span != nil && span.Tracer() != nil {
		span1 := span.Tracer().StartSpan("executor.Compile", opentracing.ChildOf(span.Context()))
		defer span1.Finish()
	}

	//TODO: infoSchema

	//Preprocess

	//Optimize

	return &ExecStmt{
		//InfoSchema
		Plan: finalPlan,
		//Expensive
		//Cacheable
		Text:     stmtNode.Text(),
		StmtNode: stmtNode,
		Ctx:      c.Ctx,
	}, nil
}
