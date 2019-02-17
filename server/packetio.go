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
// https://github.com/pingcap/tidb/blob/source-code/server/packetio.go
//

package server

import (
	"bufio"
	"io"
	"net"

	"github.com/pingcap/errors"
	"github.com/pingcap/parser/mysql"
	"github.com/pingcap/parser/terror"
)

const (
	defaultReaderSize = 16 * 1024
	defaultWriterSize = 16 * 1024
)

type packetIO struct {
	conn      net.Conn
	bufReader *bufio.Reader
	bufWriter *bufio.Writer
	sequence  uint8
}

func newPacketIO(conn net.Conn) *packetIO {
	p := &packetIO{
		conn:     conn,
		sequence: 0,
	}
	p.setBuffer(conn)
	return p
}

func (p *packetIO) setBuffer(conn net.Conn) {
	p.bufReader = bufio.NewReaderSize(conn, defaultReaderSize)
	p.bufWriter = bufio.NewWriterSize(conn, defaultWriterSize)
}

func (p *packetIO) readOnePacket() ([]byte, error) {
	var header [4]byte

	if _, err := io.ReadFull(p.conn, header[:]); err != nil {
		return nil, errors.Trace(err)
	}

	sequence := header[3]
	if sequence != p.sequence {
		return nil, errInvalidSequence.GenWithStack("invalid sequence %d != %d", sequence, p.sequence)
	}

	p.sequence++

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)

	data := make([]byte, length)
	if _, err := io.ReadFull(p.conn, data); err != nil {
		return nil, errors.Trace(err)
	}
	return data, nil
}

func (p *packetIO) readPacket() ([]byte, error) {
	data, err := p.readOnePacket()
	if err != nil {
		return nil, errors.Trace(err)
	}

	if len(data) < mysql.MaxPayloadLen {
		return data, nil
	}

	// handle multi-packet
	for {
		buf, err := p.readOnePacket()
		if err != nil {
			return nil, errors.Trace(err)
		}

		data = append(data, buf...)

		if len(buf) < mysql.MaxPayloadLen {
			break
		}
	}

	return data, nil
}

// writePacket writes data that already have header
func (p *packetIO) writePacket(data []byte) error {
	length := len(data) - 4

	for length >= mysql.MaxPayloadLen {
		data[0] = 0xff
		data[1] = 0xff
		data[2] = 0xff

		data[3] = p.sequence

		if n, err := p.bufWriter.Write(data[:4+mysql.MaxPayloadLen]); err != nil {
			terror.Log(errors.Trace(err))
			return errors.Trace(mysql.ErrBadConn)
		} else if n != (4 + mysql.MaxPayloadLen) {
			return errors.Trace(mysql.ErrBadConn)
		} else {
			p.sequence++
			length -= mysql.MaxPayloadLen
			data = data[mysql.MaxPayloadLen:]
		}
	}

	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = p.sequence

	if n, err := p.bufWriter.Write(data); err != nil {
		terror.Log(errors.Trace(err))
		return errors.Trace(mysql.ErrBadConn)
	} else if n != len(data) {
		return errors.Trace(mysql.ErrBadConn)
	} else {
		p.sequence++
		return nil
	}
}

func (p *packetIO) flush() error {
	return p.bufWriter.Flush()
}
