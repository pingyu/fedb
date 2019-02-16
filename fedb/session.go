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

package fedb

import (
	"github.com/juju/errors"

	"fedb/sessionctx/variable"
	"fedb/terror"
	"github.com/pingyu/parser/charset"
)

// Session is the session interface
type Session interface {
	SetConnectionID(uint64) Session
	SetCollation(coID int) error
	SetClientCapability(uint32) Session

	Close()
}

type session struct {
	//TODO: store
	//TODO: parser
	sessionVars *variable.SessionVars
}

var (
	_ Session = (*session)(nil)
)

// CreateSession creates a new session environment.
func CreateSession() (Session, error) {
	s := &session{
		//TODO: store
		//TODO: parser
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
