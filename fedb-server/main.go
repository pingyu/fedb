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
	var err error
	svr, err = server.NewServer(cfg)
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
