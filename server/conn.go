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
// https://github.com/pingcap/tidb/blob/source-code/server/conn.go
//

package server

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/opentracing/opentracing-go"
	goctx "golang.org/x/net/context"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
	log "github.com/sirupsen/logrus"

	"fedb/util"
	"fedb/util/arena"
	"fedb/util/hack"
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
		salt:         util.RandomBuf(20),
		alloc:        arena.NewAllocator(32 * 1024),
		status:       connStatusDispatching,
	}
}

type clientConn struct {
	pkt  *packetIO // a helper to read and write data in packet format.
	conn net.Conn  // net.Conn
	//tlsConn      *tls.Conn         // TLS connection, nil if not TLS.
	server       *Server           // a reference of server instance.
	capability   uint32            // client capability affects the way server handles client request.
	connectionID uint32            // atomically allocated by a global variable, unique in process scope.
	collation    uint8             // collation used by client, may be different from the collation used by database.
	user         string            // user of the client.
	dbname       string            // default database name.
	salt         []byte            // random bytes used for authentication.
	alloc        arena.Allocator   // an memory allocator for reducing memory allocation.
	lastCmd      string            // latest sql query string, currently used for logging error.
	ctx          QueryCtx          // an interface to execute sql statements.
	attrs        map[string]string // attributes parsed from client handshake response, not used for now.
	status       int32             // dispatching/reading/shutdown/waitshutdown

	// mu is used for cancelling the execution of current transaction.
	mu struct {
		sync.RWMutex
		cancelFunc goctx.CancelFunc
	}
}

func (cc *clientConn) setConn(conn net.Conn) {
	cc.conn = conn
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
			err1 := cc.writeError(err)
			terror.Log(errors.Trace(err1))
		}
		cc.pkt.sequence = 0
	}
}

// handshake works like TCP handshake, but in a higher level, it first writes initial packet to client,
// during handshake, client and server negotiate compatible features and do authentication.
// After handshake, client can send sql query to server.
func (cc *clientConn) handshake() error {
	var err error
	if err = cc.writeInitialHandshake(); err != nil {
		return errors.Trace(err)
	}
	if err = cc.readOptionalSSLRequestAndHandshakeResponse(); err != nil {
		err1 := cc.writeError(err)
		terror.Log(errors.Trace(err1))
		return errors.Trace(err)
	}
	data := cc.alloc.AllocWithLen(4, 32)
	data = append(data, mysql.OKHeader)
	data = append(data, 0, 0)
	if cc.capability&mysql.ClientProtocol41 > 0 {
		data = dumpUint16(data, mysql.ServerStatusAutocommit)
		data = append(data, 0, 0)
	}

	cc.ctx, err = cc.server.driver.OpenCtx(uint64(cc.connectionID), cc.capability, cc.collation, cc.dbname, nil)
	if err != nil {
		return errors.Trace(err)
	}

	err = cc.writePacket(data)
	cc.pkt.sequence = 0
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(cc.flush())
}

// writeInitialHandshake sends server version, connection ID, server capability, collation, server status
// and auth salt to the client.
func (cc *clientConn) writeInitialHandshake() error {
	data := make([]byte, 4, 128)

	// min version 10
	data = append(data, 10)
	// server version[00]
	data = append(data, mysql.ServerVersion...)
	data = append(data, 0)
	// connection id
	data = append(data, byte(cc.connectionID), byte(cc.connectionID>>8), byte(cc.connectionID>>16), byte(cc.connectionID>>24))
	// auth-plugin-data-part-1
	data = append(data, cc.salt[0:8]...)
	// filler [00]
	data = append(data, 0)
	// capability flag lower 2 bytes, using default capability here
	data = append(data, byte(cc.server.capability), byte(cc.server.capability>>8))
	// charset
	if cc.collation == 0 {
		cc.collation = uint8(mysql.DefaultCollationID)
	}
	data = append(data, cc.collation)
	// status
	data = dumpUint16(data, mysql.ServerStatusAutocommit)
	// below 13 byte may not be used
	// capability flag upper 2 bytes, using default capability here
	data = append(data, byte(cc.server.capability>>16), byte(cc.server.capability>>24))
	// length of auth-plugin-data
	data = append(data, byte(len(cc.salt)+1))
	// reserved 10 [00]
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)
	// auth-plugin-data-part-2
	data = append(data, cc.salt[8:]...)
	data = append(data, 0)
	// auth-plugin name
	data = append(data, []byte("mysql_native_password")...)
	data = append(data, 0)
	err := cc.writePacket(data)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(cc.flush())
}

type handshakeResponse41 struct {
	Capability uint32
	Collation  uint8
	User       string
	DBName     string
	Auth       []byte
	Attrs      map[string]string
}

// parseHandshakeResponseHeader parses the common header of SSLRequest and HandshakeResponse41.
func parseHandshakeResponseHeader(packet *handshakeResponse41, data []byte) (parsedBytes int, err error) {
	// Ensure there are enough data to read:
	// http://dev.mysql.com/doc/internals/en/connection-phase-packets.html#packet-Protocol::SSLRequest
	if len(data) < 4+4+1+23 {
		log.Errorf("Got malformed handshake response, packet data: %v", data)
		return 0, mysql.ErrMalformPacket
	}

	offset := 0
	// capability
	capability := binary.LittleEndian.Uint32(data[:4])
	packet.Capability = capability
	offset += 4
	// skip max packet size
	offset += 4
	// charset, skip, if you want to use another charset, use set names
	packet.Collation = data[offset]
	offset++
	// skip reserved 23[00]
	offset += 23

	return offset, nil
}

// parseHandshakeResponseBody parse the HandshakeResponse (except the common header part).
func parseHandshakeResponseBody(packet *handshakeResponse41, data []byte, offset int) (err error) {
	defer func() {
		// Check malformat packet cause out of range is disgusting, but don't panic!
		if r := recover(); r != nil {
			log.Errorf("handshake panic, packet data: %v", data)
			err = mysql.ErrMalformPacket
		}
	}()
	// user name
	packet.User = string(data[offset : offset+bytes.IndexByte(data[offset:], 0)])
	offset += len(packet.User) + 1

	if packet.Capability&mysql.ClientPluginAuthLenencClientData > 0 {
		// MySQL client sets the wrong capability, it will set this bit even server doesn't
		// support ClientPluginAuthLenencClientData.
		// https://github.com/mysql/mysql-server/blob/5.7/sql-common/client.c#L3478
		num, null, off := parseLengthEncodedInt(data[offset:])
		offset += off
		if !null {
			packet.Auth = data[offset : offset+int(num)]
			offset += int(num)
		}
	} else if packet.Capability&mysql.ClientSecureConnection > 0 {
		// auth length and auth
		authLen := int(data[offset])
		offset++
		packet.Auth = data[offset : offset+authLen]
		offset += authLen
	} else {
		packet.Auth = data[offset : offset+bytes.IndexByte(data[offset:], 0)]
		offset += len(packet.Auth) + 1
	}

	if packet.Capability&mysql.ClientConnectWithDB > 0 {
		if len(data[offset:]) > 0 {
			idx := bytes.IndexByte(data[offset:], 0)
			packet.DBName = string(data[offset : offset+idx])
			offset = offset + idx + 1
		}
	}

	if packet.Capability&mysql.ClientPluginAuth > 0 {
		// TODO: Support mysql.ClientPluginAuth, skip it now
		idx := bytes.IndexByte(data[offset:], 0)
		offset = offset + idx + 1
	}

	if packet.Capability&mysql.ClientConnectAtts > 0 {
		if len(data[offset:]) == 0 {
			// Defend some ill-formated packet, connection attribute is not important and can be ignored.
			return nil
		}
		if num, null, off := parseLengthEncodedInt(data[offset:]); !null {
			offset += off
			row := data[offset : offset+int(num)]
			attrs, err := parseAttrs(row)
			if err != nil {
				log.Warn("parse attrs error:", errors.ErrorStack(err))
				return nil
			}
			packet.Attrs = attrs
		}
	}

	return nil
}

func parseAttrs(data []byte) (map[string]string, error) {
	attrs := make(map[string]string)
	pos := 0
	for pos < len(data) {
		key, _, off, err := parseLengthEncodedBytes(data[pos:])
		if err != nil {
			return attrs, errors.Trace(err)
		}
		pos += off
		value, _, off, err := parseLengthEncodedBytes(data[pos:])
		if err != nil {
			return attrs, errors.Trace(err)
		}
		pos += off

		attrs[string(key)] = string(value)
	}
	return attrs, nil
}

func (cc *clientConn) readOptionalSSLRequestAndHandshakeResponse() error {
	// Read a packet. It may be a SSLRequest or HandshakeResponse.
	data, err := cc.readPacket()
	if err != nil {
		return errors.Trace(err)
	}

	var resp handshakeResponse41

	pos, err := parseHandshakeResponseHeader(&resp, data)
	if err != nil {
		return errors.Trace(err)
	}

	// if (resp.Capability&mysql.ClientSSL > 0) && cc.server.tlsConfig != nil {
	// 	// The packet is a SSLRequest, let's switch to TLS.
	// 	if err = cc.upgradeToTLS(cc.server.tlsConfig); err != nil {
	// 		return errors.Trace(err)
	// 	}
	// 	// Read the following HandshakeResponse packet.
	// 	data, err = cc.readPacket()
	// 	if err != nil {
	// 		return errors.Trace(err)
	// 	}
	// 	pos, err = parseHandshakeResponseHeader(&resp, data)
	// 	if err != nil {
	// 		return errors.Trace(err)
	// 	}
	// }

	// Read the remaining part of the packet.
	if err = parseHandshakeResponseBody(&resp, data, pos); err != nil {
		return errors.Trace(err)
	}

	cc.capability = resp.Capability & cc.server.capability
	cc.user = resp.User
	cc.dbname = resp.DBName
	cc.collation = resp.Collation
	cc.attrs = resp.Attrs

	// Open session and do auth.
	// var tlsStatePtr *tls.ConnectionState
	// if cc.tlsConn != nil {
	// 	tlsState := cc.tlsConn.ConnectionState()
	// 	tlsStatePtr = &tlsState
	// }
	// cc.ctx, err = cc.server.driver.OpenCtx(uint64(cc.connectionID), cc.capability, cc.collation, cc.dbname, nil)
	// if err != nil {
	// 	return errors.Trace(err)
	// }
	// if !cc.server.skipAuth() {
	// 	// Do Auth.
	// 	addr := cc.bufReadConn.RemoteAddr().String()
	// 	host, _, err1 := net.SplitHostPort(addr)
	// 	if err1 != nil {
	// 		return errors.Trace(errAccessDenied.GenByArgs(cc.user, addr, "YES"))
	// 	}
	// 	if !cc.ctx.Auth(&auth.UserIdentity{Username: cc.user, Hostname: host}, resp.Auth, cc.salt) {
	// 		return errors.Trace(errAccessDenied.GenByArgs(cc.user, host, "YES"))
	// 	}
	// }
	// if cc.dbname != "" {
	// 	err = cc.useDB(goctx.Background(), cc.dbname)
	// 	if err != nil {
	// 		return errors.Trace(err)
	// 	}
	// }
	// cc.ctx.SetSessionManager(cc.server)
	// if cc.server.cfg.EnableChunk {
	// 	cc.ctx.EnableChunk()
	// }
	return nil
}

// dispatch handles client request based on command which is the first byte of the data.
// It also gets a token from server which is used to limit the concurrently handling clients.
// The most frequently used command is ComQuery.
func (cc *clientConn) dispatch(data []byte) error {
	span := opentracing.StartSpan("server.dispatch")
	goCtx := opentracing.ContextWithSpan(goctx.Background(), span)

	goCtx1, cancelFunc := goctx.WithCancel(goCtx)
	cc.mu.Lock()
	cc.mu.cancelFunc = cancelFunc
	cc.mu.Unlock()

	cmd := data[0]
	data = data[1:]
	cc.lastCmd = hack.String(data)
	//token := cc.server.getToken()
	defer func() {
		//cc.server.releaseToken(token)
		span.Finish()
	}()

	log.Infof("cmd:0x%x, %v", cmd, data)

	switch cmd {
	case mysql.ComSleep:
		// TODO: According to mysql document, this command is supposed to be used only internally.
		// So it's just a temp fix, not sure if it's done right.
		// Investigate this command and write test case later.
		return nil
	case mysql.ComQuit:
		return io.EOF
	case mysql.ComQuery: // Most frequently used command.
		// For issue 1989
		// Input payload may end with byte '\0', we didn't find related mysql document about it, but mysql
		// implementation accept that case. So trim the last '\0' here as if the payload an EOF string.
		// See http://dev.mysql.com/doc/internals/en/com-query.html
		if len(data) > 0 && data[len(data)-1] == 0 {
			data = data[:len(data)-1]
		}
		return cc.handleQuery(goCtx1, hack.String(data))
	case mysql.ComPing:
		return cc.writeOK()
	case mysql.ComInitDB:
		if err := cc.useDB(goCtx1, hack.String(data)); err != nil {
			return errors.Trace(err)
		}
		return cc.writeOK()
	// case mysql.ComFieldList:
	// 	return cc.handleFieldList(hack.String(data))
	// case mysql.ComStmtPrepare:
	// 	return cc.handleStmtPrepare(hack.String(data))
	// case mysql.ComStmtExecute:
	// 	return cc.handleStmtExecute(goCtx1, data)
	// case mysql.ComStmtClose:
	// 	return cc.handleStmtClose(data)
	// case mysql.ComStmtSendLongData:
	// 	return cc.handleStmtSendLongData(data)
	// case mysql.ComStmtReset:
	// 	return cc.handleStmtReset(data)
	// case mysql.ComSetOption:
	// 	return cc.handleSetOption(data)
	default:
		return mysql.NewErrf(mysql.ErrUnknown, "command %d not supported now", cmd)
	}
}

func (cc *clientConn) writeOK() error {
	data := cc.alloc.AllocWithLen(4, 32)
	data = append(data, mysql.OKHeader)
	//TODO data = dumpLengthEncodedInt(data, cc.ctx.AffectedRows())
	data = dumpLengthEncodedInt(data, 0)
	//TODO data = dumpLengthEncodedInt(data, cc.ctx.LastInsertID())
	data = dumpLengthEncodedInt(data, 0)
	if cc.capability&mysql.ClientProtocol41 > 0 {
		//TODO data = dumpUint16(data, cc.ctx.Status())
		data = dumpUint16(data, mysql.ServerStatusAutocommit)
		//TODO data = dumpUint16(data, cc.ctx.WarningCount())
		data = dumpUint16(data, 0)
	}

	err := cc.writePacket(data)
	if err != nil {
		return errors.Trace(err)
	}

	return errors.Trace(cc.flush())
}

func (cc *clientConn) writeError(e error) error {
	var (
		m  *mysql.SQLError
		te *terror.Error
		ok bool
	)
	originErr := errors.Cause(e)
	if te, ok = originErr.(*terror.Error); ok {
		m = te.ToSQLError()
	} else {
		m = mysql.NewErrf(mysql.ErrUnknown, "%s", e.Error())
	}

	data := cc.alloc.AllocWithLen(4, 16+len(m.Message))
	data = append(data, mysql.ErrHeader)
	data = append(data, byte(m.Code), byte(m.Code>>8))
	if cc.capability&mysql.ClientProtocol41 > 0 {
		data = append(data, '#')
		data = append(data, m.State...)
	}

	data = append(data, m.Message...)

	err := cc.writePacket(data)
	if err != nil {
		return errors.Trace(err)
	}
	return errors.Trace(cc.flush())
}

func (cc *clientConn) writePacket(data []byte) error {
	return cc.pkt.writePacket(data)
}

func (cc *clientConn) flush() error {
	return cc.pkt.flush()
}

func (cc *clientConn) Close() error {
	cc.server.rwlock.Lock()
	delete(cc.server.clients, cc.connectionID)
	cc.server.rwlock.Unlock()

	err := cc.conn.Close()
	terror.Log(errors.Trace(err))
	if cc.ctx != nil {
		return cc.ctx.Close()
	}
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

func (cc *clientConn) useDB(goCtx goctx.Context, db string) (err error) {
	//TODO change DB
	cc.dbname = db
	return nil
}

func (cc *clientConn) handleQuery(goCtx goctx.Context, sql string) (err error) {
	_, err = cc.ctx.Execute(goCtx, sql)
	//TODO

	err = cc.writeOK()
	return errors.Trace(err)
}
