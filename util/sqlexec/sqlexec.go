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
// https://github.com/pingcap/tidb/blob/master/util/sqlexec/restricted_sql_executor.go

package sqlexec

// RecordSet is an abstract result set interface to help get data from Plan.
type RecordSet interface {
	// Fields gets result fields.
	//Fields() []*ast.ResultField

	// Next reads records into chunk.
	//Next(ctx context.Context, req *chunk.RecordBatch) error

	//NewRecordBatch create a recordBatch.
	//NewRecordBatch() *chunk.RecordBatch

	// Close closes the underlying iterator, call Next after Close will
	// restart the iteration.
	//Close() error
}
