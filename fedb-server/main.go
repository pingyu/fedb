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
	"fedb/config"
	"fedb/server"
	"fedb/terror"
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	cfg      *config.Config
	svr      *server.Server
	graceful bool
)

func main() {
	fmt.Println("Hello, FeDB !!")

	loadConfig()
	createServer()
	setupSignalHandler()
	runServer()
	cleanup()
	os.Exit(0)
}

func loadConfig() {
	cfg = config.GetGlobalConfig()
}

func createServer() {
	var driver server.IDriver
	driver = server.NewFeDBDriver()
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
