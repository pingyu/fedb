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
//https://github.com/pingcap/tidb/blob/master/util/misc.go

package util

import (
	log "github.com/sirupsen/logrus"

	"github.com/pingcap/parser"
)

const (
	// syntaxErrorPrefix is the common prefix for SQL syntax error in TiDB.
	syntaxErrorPrefix = "You have an error in your SQL syntax; check the manual that corresponds to your FeDB version for the right syntax to use"
)

// SyntaxError converts parser error to TiDB's syntax error.
func SyntaxError(err error) error {
	if err == nil {
		return nil
	}
	log.Errorf("%+v", err)
	return parser.ErrParse.GenWithStackByArgs(syntaxErrorPrefix, err.Error())
}
