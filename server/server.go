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
// https://github.com/pingcap/tidb/blob/source-code/server/server.go
//

package server

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
	log "github.com/sirupsen/logrus"

	"fedb/config"
)

// Server error codes.
const (
	codeUnknownFieldType  = 1
	codeInvalidPayloadLen = 2
	codeInvalidSequence   = 3
	codeInvalidType       = 4

	codeNotAllowedCommand = 1148
	codeAccessDenied      = mysql.ErrAccessDenied
)

var (
	baseConnID uint32
)

var (
	//errUnknownFieldType  = terror.ClassServer.New(codeUnknownFieldType, "unknown field type")
	//errInvalidPayloadLen = terror.ClassServer.New(codeInvalidPayloadLen, "invalid payload length")
	errInvalidSequence = terror.ClassServer.New(codeInvalidSequence, "invalid sequence")
	//errInvalidType       = terror.ClassServer.New(codeInvalidType, "invalid type")
	//errNotAllowedCommand = terror.ClassServer.New(codeNotAllowedCommand, "the used command is not allowed with this TiDB version")
	//errAccessDenied      = terror.ClassServer.New(codeAccessDenied, mysql.MySQLErrName[mysql.ErrAccessDenied])
)

// DefaultCapability is the capability of the server when it is created using the default configuration.
// When server is configured with SSL, the server will have extra capabilities compared to DefaultCapability.
const defaultCapability = mysql.ClientLongPassword | mysql.ClientLongFlag |
	mysql.ClientConnectWithDB | mysql.ClientProtocol41 |
	mysql.ClientTransactions | mysql.ClientSecureConnection | mysql.ClientFoundRows |
	mysql.ClientMultiStatements | mysql.ClientMultiResults | mysql.ClientLocalFiles |
	mysql.ClientConnectAtts | mysql.ClientPluginAuth

// Server is the MySQL protocol server
type Server struct {
	cfg *config.Config
	//tlsConfig         *tls.Config
	driver   IDriver
	listener net.Listener
	rwlock   *sync.RWMutex
	//concurrentLimiter *TokenLimiter
	clients    map[uint32]*clientConn
	capability uint32

	// stopListenerCh is used when a critical error occurred, we don't want to exit the process, because there may be
	// a supervisor automatically restart it, then new client connection will be created, but we can't server it.
	// So we just stop the listener and store to force clients to chose other TiDB servers.
	//stopListenerCh chan struct{}
}

// NewServer create server
func NewServer(cfg *config.Config, driver IDriver) (*Server, error) {
	s := &Server{
		cfg:    cfg,
		driver: driver,
		//concurrentLimiter
		rwlock:  &sync.RWMutex{},
		clients: make(map[uint32]*clientConn),
	}

	s.capability = defaultCapability
	// tlsConfig

	var err error
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	if s.listener, err = net.Listen("tcp", addr); err == nil {
		log.Infof("Server listen at [%s]", addr)
	}

	if err != nil {
		return nil, errors.Trace(err)
	}

	rand.Seed(time.Now().UTC().UnixNano())
	return s, nil
}

// Close close server
func (s *Server) Close() {
	s.rwlock.Lock()
	defer s.rwlock.Unlock()

	if s.listener != nil {
		err := s.listener.Close()
		terror.Log(errors.Trace(err))
		s.listener = nil
	}
}

// Run server
func (s *Server) Run() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if opErr, ok := err.(*net.OpError); ok {
				if opErr.Err.Error() == "use of closed network connection" {
					log.Infof("Normal quit of Server.Run()")
					return nil
				}
			}

			log.Errorf("accept error %s", err.Error())
			return errors.Trace(err)
		}
		go s.onConn(conn)
	}
}

func (s *Server) onConn(c net.Conn) {
	conn := s.newConn(c)
	defer func() {
		log.Infof("[%d] close connection", conn.connectionID)
	}()

	if err := conn.handshake(); err != nil {
		log.Infof("handshake error %s", errors.ErrorStack(err))
		err = c.Close()
		terror.Log(errors.Trace(err))
		return
	}

	s.rwlock.Lock()
	s.clients[conn.connectionID] = conn
	s.rwlock.Unlock()

	conn.Run()
}

func (s *Server) newConn(conn net.Conn) *clientConn {
	cc := newClientConn(s)
	log.Infof("[%d] new connection %s", cc.connectionID, conn.RemoteAddr().String())
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetKeepAlive(true); err != nil {
			log.Error("failed to set tcp keep alive option:", err)
		}
	}
	cc.setConn(conn)
	return cc
}

// GracefulDown graceful shutdown the server
func (s *Server) GracefulDown() {
	log.Infof("[server] graceful shutdown.")

	//TODO
}
