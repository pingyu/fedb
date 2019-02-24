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
// https://github.com/pingcap/tidb/blob/source-code/sessionctx/variable/session.go
//

package variable

import (
	"time"
)

// SessionVars is session variables
type SessionVars struct {
	systems map[string]string // systems variables

	// Should be reset on transaction finished.
	TxnCtx *TransactionContext

	// Following variables are special for current session.
	Status uint16

	ClientCapability uint32 // ClientCapability is client capability
	ConnectionID     uint64 // ConnectionID is connection id
	CurrentDB        string // CurrentDB is current db name
}

// NewSessionVars create SessionVars
func NewSessionVars() *SessionVars {
	return &SessionVars{
		systems: make(map[string]string),
		TxnCtx:  &TransactionContext{},
	}
}

// SetSystemVar sets the value of system variable.
func (s *SessionVars) SetSystemVar(name string, val string) error {
	s.systems[name] = val
	return nil
}

// GetSystemVar gets value of system variable.
func (s *SessionVars) GetSystemVar(name string) (string, bool) {
	val, ok := s.systems[name]
	return val, ok
}

// GetCharsetInfo gets charset and collation for current context.
// What character set should the server translate a statement to after receiving it?
// For this, the server uses the character_set_connection and collation_connection system variables.
// It converts statements sent by the client from character_set_client to character_set_connection
// (except for string literals that have an introducer such as _latin1 or _utf8).
// collation_connection is important for comparisons of literal strings.
// For comparisons of strings with column values, collation_connection does not matter because columns
// have their own collation, which has a higher collation precedence.
// See https://dev.mysql.com/doc/refman/5.7/en/charset-connection.html
func (s *SessionVars) GetCharsetInfo() (charset, collation string) {
	charset = s.systems[CharacterSetConnection]
	collation = s.systems[CollationConnection]
	return
}

// TransactionContext is used to store variables that has transaction scope.
type TransactionContext struct {
	//ForUpdate     bool
	//DirtyDB       interface{}
	//Binlog        interface{}
	//InfoSchema    interface{}
	//History       interface{}
	//SchemaVersion int64
	StartTS uint64
	//Shard         *int64
	//TableDeltaMap map[int64]TableDelta

	// For metrics.
	CreateTime time.Time
	//StatementCount int
}
