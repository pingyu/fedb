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

// SessionVars is session variables
type SessionVars struct {
	systems map[string]string // systems variables

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
