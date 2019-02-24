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
//https://github.com/pingcap/tidb/blob/master/planner/core/preprocess.go

package core

import (
	"github.com/pingcap/errors"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/model"
	log "github.com/sirupsen/logrus"

	"fedb/sessionctx"
)

// Preprocess resolves table names of the node
func Preprocess(ctx sessionctx.Context, node ast.Node) error {
	v := preprocessor{
		//is
		ctx:              ctx,
		tableAliasInJoin: make([]map[string]interface{}, 0),
	}
	node.Accept(&v)
	return errors.Trace(v.err)
}

type preprocessor struct {
	//is
	ctx sessionctx.Context
	err error
	//inPrepare

	// inCreateOrDropTable is true when visiting create/drop table statement.
	inCreateOrDropTable bool

	// tableAliasInJoin is a stack that keeps the table alias names for joins.
	// len(tableAliasInJoin) may bigger than 1 because the left/right child of join may be subquery that contains `JOIN`
	tableAliasInJoin []map[string]interface{}

	parentIsJoin bool
}

func (p *preprocessor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	switch in.(type) {
	case *ast.Join:
		//p.checkNonUniqTableAlias(node)
		p.parentIsJoin = true
	default:
		p.parentIsJoin = false
	}
	return in, p.err != nil
}

func (p *preprocessor) Leave(in ast.Node) (out ast.Node, ok bool) {
	switch x := in.(type) {
	case *ast.TableName:
		p.handleTableName(x)
	case *ast.Join:
		//TODO
	}

	return in, p.err != nil
}

func (p *preprocessor) handleTableName(tn *ast.TableName) {
	if tn.Schema.L == "" {
		currentDB := p.ctx.GetSessionVars().CurrentDB
		if currentDB == "" {
			p.err = errors.Trace(ErrNoDB)
			return
		}
		tn.Schema = model.NewCIStr(currentDB)
	}
	log.Infof("handleTableName: %s.%s", tn.Schema.L, tn.Name.L)
}
