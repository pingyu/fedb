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
//https://github.com/pingcap/tidb/blob/master/kv/kv.go

package kv

import (
	"fedb/store/fekv/oracle"
)

// Transaction defines the interface for operations inside a Transaction.
// This is not thread safe.
type Transaction interface {
	//MemBuffer
	// Commit commits the transaction operations to KV store.
	//Commit(context.Context) error
	// Rollback undoes the transaction operations to KV store.
	//Rollback() error
	// String implements fmt.Stringer interface.
	//String() string
	// LockKeys tries to lock the entries with the keys in KV store.
	//LockKeys(keys ...Key) error
	// SetOption sets an option with a value, when val is nil, uses the default
	// value of this option.
	//SetOption(opt Option, val interface{})
	// DelOption deletes an option.
	//DelOption(opt Option)
	// IsReadOnly checks if the transaction has only performed read operations.
	//IsReadOnly() bool
	// StartTS returns the transaction start timestamp.
	//StartTS() uint64
	// Valid returns if the transaction is valid.
	// A transaction become invalid after commit or rollback.
	Valid() bool
	// GetMemBuffer return the MemBuffer binding to this transaction.
	//GetMemBuffer() MemBuffer
	// GetSnapshot returns the snapshot of this transaction.
	//GetSnapshot() Snapshot
	// SetVars sets variables to the transaction.
	//SetVars(vars *Variables)
}

// Driver is the interface that must be implemented by a KV storage.
type Driver interface {
	// Open returns a new Storage.
	// The path is the string for storage specific format.
	Open(path string) (Storage, error)
}

// Storage defines the interface for storage.
// Isolation should be at least SI(SNAPSHOT ISOLATION)
type Storage interface {
	// Begin transaction
	//Begin() (Transaction, error)
	// BeginWithStartTS begins transaction with startTS.
	//BeginWithStartTS(startTS uint64) (Transaction, error)
	// GetSnapshot gets a snapshot that is able to read any data which data is <= ver.
	// if ver is MaxVersion or > current max committed version, we will use current version for this snapshot.
	//GetSnapshot(ver Version) (Snapshot, error)
	// GetClient gets a client instance.
	//GetClient() Client
	// Close store
	//Close() error
	// UUID return a unique ID which represents a Storage.
	//UUID() string
	// CurrentVersion returns current max committed version.
	//CurrentVersion() (Version, error)
	// GetOracle gets a timestamp oracle client.
	GetOracle() oracle.Oracle
	// SupportDeleteRange gets the storage support delete range or not.
	//SupportDeleteRange() (supported bool)
}
