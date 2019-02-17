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
// https://github.com/pingcap/tidb/blob/source-code/session.go
//

package session

import (
	"fmt"

	"github.com/opentracing/opentracing-go"
	goctx "golang.org/x/net/context"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser"
	"github.com/pingcap/parser/ast"
	"github.com/pingcap/parser/charset"
	"github.com/pingcap/parser/terror"
	log "github.com/sirupsen/logrus"

	"fedb/sessionctx/variable"
	"fedb/util/sqlexec"
)

// Session is the session interface
type Session interface {
	Execute(goctx.Context, string) ([]sqlexec.RecordSet, error) // Execute a sql statement.

	SetConnectionID(uint64) Session
	SetCollation(coID int) error
	SetClientCapability(uint32) Session

	Close()
}

type session struct {
	//TODO: store
	parser      *parser.Parser
	sessionVars *variable.SessionVars
}

var (
	_ Session = (*session)(nil)
)

// CreateSession creates a new session environment.
func CreateSession() (Session, error) {
	s := &session{
		//TODO: store
		parser:      parser.New(),
		sessionVars: variable.NewSessionVars(),
	}
	return s, nil
}

func (s *session) SetConnectionID(connectionID uint64) Session {
	s.sessionVars.ConnectionID = connectionID
	return s
}

func (s *session) SetClientCapability(capability uint32) Session {
	s.sessionVars.ClientCapability = capability
	return s
}

func (s *session) SetCollation(coID int) error {
	cs, co, err := charset.GetCharsetInfoByID(coID)
	if err != nil {
		return errors.Trace(err)
	}
	for _, v := range variable.SetNamesVariables {
		terror.Log(errors.Trace(s.sessionVars.SetSystemVar(v, cs)))
	}
	terror.Log(errors.Trace(s.sessionVars.SetSystemVar(variable.CollationConnection, co)))
	return nil
}

func (s *session) Close() {
	// statsCollector
	// RoolbackTxn
}

type visitor struct{}

func (v *visitor) Enter(in ast.Node) (out ast.Node, skipChildren bool) {
	fmt.Printf("Enter %T\n", in)
	return in, false
}

func (v *visitor) Leave(in ast.Node) (out ast.Node, ok bool) {
	fmt.Printf("Leave: %T\n", in)
	return in, true
}

// Execute a sql statement.
func (s *session) Execute(ctx goctx.Context, sql string) (recordSets []sqlexec.RecordSet, err error) {
	log.Infof("sql: %v", sql)

	if span := opentracing.SpanFromContext(ctx); span != nil && span.Tracer() != nil {
		span1 := span.Tracer().StartSpan("session.Execute", opentracing.ChildOf(span.Context()))
		defer span1.Finish()
	}
	recordSets, err = s.execute(ctx, sql)
	return
}

func (s *session) execute(ctx goctx.Context, sql string) (recordSets []sqlexec.RecordSet, err error) {
	charsetInfo, collation := s.sessionVars.GetCharsetInfo()

	stmtNodes, err := s.parser.Parse(sql, charsetInfo, collation)
	if err != nil {
		return nil, errors.AddStack(err)
	}

	for _, stmtNode := range stmtNodes {
		v := visitor{}
		stmtNode.Accept(&v)
	}
	//TODO
	//compiler

	return nil, nil
}
