package server

import (
	"fedb/config"
	"fedb/terror"
	"fmt"
	"net"
	"sync"

	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"fedb/mysql"
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

// Server object
type Server struct {
	cfg      *config.Config
	listener net.Listener
	rwlock   *sync.RWMutex
	clients  map[uint32]*clientConn
}

// NewServer create server
func NewServer(cfg *config.Config) (*Server, error) {
	s := &Server{
		cfg: cfg,
		//driver
		//concurrentLimiter
		rwlock:  &sync.RWMutex{},
		clients: make(map[uint32]*clientConn),
	}

	//capability

	var err error
	addr := fmt.Sprintf("%s:%d", s.cfg.Host, s.cfg.Port)
	if s.listener, err = net.Listen("tcp", addr); err == nil {
		log.Infof("Server listen at [%s]", addr)
	}

	if err != nil {
		return nil, errors.Trace(err)
	}

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
