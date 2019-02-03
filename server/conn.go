package server

import (
	"fedb/mysql"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	goctx "github.com/golang/net/context"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"

	"fedb/terror"
	"fedb/util/arena"
)

const (
	connStatusDispatching int32 = iota
	connStatusReading
	connStatusShutdown     // Closed by server.
	connStatusWaitShutdown // Notified by server to close.
)

func newClientConn(s *Server) *clientConn {
	return &clientConn{
		server:       s,
		connectionID: atomic.AddUint32(&baseConnID, 1),
		collation:    mysql.DefaultCollationID,
		alloc:        arena.NewAllocator(32 * 1024),
		status:       connStatusDispatching,
	}
}

type clientConn struct {
	pkt  *packetIO // a helper to read and write data in packet format.
	conn net.Conn  // net.Conn
	//tlsConn      *tls.Conn         // TLS connection, nil if not TLS.
	server *Server // a reference of server instance.
	//capability   uint32            // client capability affects the way server handles client request.
	connectionID uint32 // atomically allocated by a global variable, unique in process scope.
	collation    uint8  // collation used by client, may be different from the collation used by database.
	user         string // user of the client.
	dbname       string // default database name.
	//salt         []byte            // random bytes used for authentication.
	alloc   arena.Allocator // an memory allocator for reducing memory allocation.
	lastCmd string          // latest sql query string, currently used for logging error.
	//ctx          QueryCtx          // an interface to execute sql statements.
	//attrs        map[string]string // attributes parsed from client handshake response, not used for now.
	status int32 // dispatching/reading/shutdown/waitshutdown

	// mu is used for cancelling the execution of current transaction.
	mu struct {
		sync.RWMutex
		cancelFunc goctx.CancelFunc
	}
}

func (cc *clientConn) setConn(conn net.Conn) {
	if cc.pkt == nil {
		cc.pkt = newPacketIO(conn)
	} else {
		// Preserve current sequence number.
		cc.pkt.setBuffer(conn)
	}
}

func (cc *clientConn) readPacket() ([]byte, error) {
	return cc.pkt.readPacket()
}

func (cc *clientConn) Run() {
	closedOutside := false
	defer func() {
		r := recover()
		if r != nil {
			buf := make([]byte, 4096)
			stackSize := runtime.Stack(buf, false)
			buf = buf[:stackSize]
			log.Errorf("lastCmd %s, %v, %s", cc.lastCmd, r, buf)
		}
		if !closedOutside {
			err := cc.Close()
			terror.Log(errors.Trace(err))
		}
	}()

	for {
		if atomic.CompareAndSwapInt32(&cc.status, connStatusDispatching, connStatusReading) == false {
			if atomic.LoadInt32(&cc.status) == connStatusShutdown {
				closedOutside = true
			}
			return
		}

		cc.alloc.Reset()
		data, err := cc.readPacket()
		if err != nil {
			if terror.ErrorNotEqual(err, io.EOF) {
				errStack := errors.ErrorStack(err)
				if !strings.Contains(errStack, "use of closed network connection") {
					log.Errorf("[%d] read packet error, close this connection %s",
						cc.connectionID, errStack)
				}
			}
			return
		}

		if atomic.CompareAndSwapInt32(&cc.status, connStatusReading, connStatusDispatching) == false {
			if atomic.LoadInt32(&cc.status) == connStatusShutdown {
				closedOutside = true
			}
			return
		}

		//startTime := time.Now()
		if err = cc.dispatch(data); err != nil {
			if terror.ErrorEqual(err, io.EOF) {
				return
			} else if terror.ErrResultUndetermined.Equal(err) {
				log.Errorf("[%d] result undetermined error, close this connection %s",
					cc.connectionID, errors.ErrorStack(err))
				return
			} else if terror.ErrCritical.Equal(err) {
				log.Errorf("[%d] critical error, stop the server listener %s",
					cc.connectionID, errors.ErrorStack(err))
				// stopListenerCh
				return
			}
			log.Warnf("[%d] dispatch error:\n%s\n%q\n%s",
				cc.connectionID, cc, queryStrForLog(string(data[1:])), errStrForLog(err))
			//err1 := cc.writeError(err)
			//terror.Log(errors.Trace(err1))
		}
		cc.pkt.sequence = 0
	}
}

func (cc *clientConn) dispatch(data []byte) error {
	//TODO
	return nil
}

func (cc *clientConn) Close() error {
	cc.server.rwlock.Lock()
	delete(cc.server.clients, cc.connectionID)
	cc.server.rwlock.Unlock()

	err := cc.conn.Close()
	terror.Log(errors.Trace(err))
	// if cc.ctx != nil {
	// 	return cc.ctx.Close()
	// }
	return nil
}

func (cc *clientConn) String() string {
	collationStr := mysql.Collations[cc.collation]
	return fmt.Sprintf("id:%d, addr:%s status:%d, collation:%s, user:%s",
		cc.connectionID, cc.conn.RemoteAddr(), cc.status /*cc.ctx.Status()*/, collationStr, cc.user,
	)
}

func queryStrForLog(query string) string {
	const size = 4096
	if len(query) > size {
		return query[:size] + fmt.Sprintf("(len: %d)", len(query))
	}
	return query
}

func errStrForLog(err error) string {
	//TODO do not log stack for duplicated entry error.
	return errors.ErrorStack(err)
}
