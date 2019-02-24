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
//https://github.com/pingcap/tidb/blob/master/store/store.go

package store

import (
	"net/url"
	"strings"

	"github.com/pingcap/errors"

	"fedb/kv"
)

var stores = make(map[string]kv.Driver)

// Register registers kv storage with name
func Register(name string, driver kv.Driver) error {
	name = strings.ToLower(name)

	if _, ok := stores[name]; ok {
		return errors.Errorf("%s is already exists", name)
	}

	stores[name] = driver
	return nil
}

const maxRetries = 30

// New creates a kv Storage with path.
//
// The path must be a URL format 'engine://path?params' like the one for
// session.Open() but with the dbname cut off.
// Examples:
//    goleveldb://relative/path
//    boltdb:///absolute/path
//
// The engine should be registered before creating storage.
func New(path string) (kv.Storage, error) {
	return newStoreWithRetry(path, maxRetries)
}

func newStoreWithRetry(path string, retries int) (kv.Storage, error) {
	url, err := url.Parse(path)
	if err != nil {
		return nil, errors.Trace(err)
	}

	name := strings.ToLower(url.Scheme)
	drv, ok := stores[name]
	if !ok {
		return nil, errors.Errorf("invalid uri, [%s] scheme not registered", name)
	}

	// retries
	s, err := drv.Open(path)
	return s, errors.Trace(err)
}
