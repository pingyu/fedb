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
// https://github.com/pingcap/tidb/blob/source-code/tidb-server/main.go
//

package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pingcap/parser/terror"
	log "github.com/sirupsen/logrus"

	"fedb/config"
	"fedb/kv"
	"fedb/server"
	kvstore "fedb/store"
	"fedb/store/localstore"

	_ "github.com/pingcap/tidb/types/parser_driver"
)

// Flag Names
const (
	nmStore     = "store"
	nmStorePath = "path"
)

var (
	store     = flag.String(nmStore, "local", "registered store name, [local, cluster]")
	storePath = flag.String(nmStorePath, "/tmp/fedb", "fedb storage path")
)

var (
	cfg      *config.Config
	storage  kv.Storage
	svr      *server.Server
	graceful bool
)

func main() {
	fmt.Println("Hello, FeDB !!")

	flag.Parse()

	registerStores()

	loadConfig()
	overrideConfig()

	createStore()
	createServer()
	setupSignalHandler()
	runServer()
	cleanup()
	os.Exit(0)
}

func loadConfig() {
	cfg = config.GetGlobalConfig()
}

func overrideConfig() {
	actualFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		actualFlags[f.Name] = true
	})
	if actualFlags[nmStore] {
		cfg.Store = *store
	}
	if actualFlags[nmStorePath] {
		cfg.Path = *storePath
	}
}

func registerStores() {
	err := kvstore.Register("local", localstore.Driver{})
	terror.MustNil(err)
}

func createStore() {
	fullPath := fmt.Sprintf("%s://%s", cfg.Store, cfg.Path)
	var err error
	storage, err = kvstore.New(fullPath)
	terror.MustNil(err)

	//BootstrapSession(storage), getStoreBootstrapVersion
}

func createServer() {
	var driver server.IDriver
	driver = server.NewFeDBDriver(storage)
	var err error
	svr, err = server.NewServer(cfg, driver)
	terror.MustNil(err)
}

func setupSignalHandler() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT,
	)

	go func() {
		sig := <-sc
		log.Infof("Got signal [%s] to exit.", sig)
		if sig == syscall.SIGTERM {
			graceful = true
		}

		svr.Close()
	}()
}

func runServer() {
	err := svr.Run()
	terror.MustNil(err)
}

func cleanup() {
	if graceful {
		svr.GracefulDown()
	}
	//TODO storage.Close
}
