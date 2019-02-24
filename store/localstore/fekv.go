//
// Copyright 2018 PingCAP, Inc.
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
//https://github.com/pingcap/tidb/blob/master/store/mockstore/tikv.go

package localstore

import (
	"net/url"
	"strings"

	"github.com/pingcap/errors"

	"fedb/kv"
	"fedb/store/fekv/oracle"
	"fedb/store/fekv/oracle/oracles"
)

// Driver is local fedb driver.
type Driver struct {
}

// Open creates a LocalKV storage.
func (drv Driver) Open(path string) (kv.Storage, error) {
	u, err := url.Parse(path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	if !strings.EqualFold(u.Scheme, "local") {
		return nil, errors.Errorf("Uri scheme expected(local) but found (%s)", u.Scheme)
	}

	opts := []localStoreOption{withPath(u.Path)}
	//txnLocalLatches
	return newLocalStore(opts...)
}

type localOptions struct {
	//mvccStore
	//clientHijack
	//pdClientHijack
	path string
	//txnLocalLatches uint
}

// LocalStoreOption control behavior of local kv
type localStoreOption func(*localOptions)

// WithPath specified local path
func withPath(path string) localStoreOption {
	return func(opt *localOptions) {
		opt.path = path
	}
}

// LocalStore is the local storage, implement kv.Storage
type LocalStore struct {
	//mvccStore
	oracle oracle.Oracle
	//safepoint
	options *localOptions
}

var _ kv.Storage = (*LocalStore)(nil)

// NewLocalStore creates local store
func newLocalStore(optionList ...localStoreOption) (kv.Storage, error) {
	var opts localOptions
	for _, f := range optionList {
		f(&opts)
	}

	//TODO: mvccStore

	oracle := oracles.NewLocalOracle()

	return &LocalStore{
		oracle:  oracle,
		options: &opts,
	}, nil
}

// GetOracle implement interface
func (s *LocalStore) GetOracle() oracle.Oracle {
	return s.oracle
}
