package protocol

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
)

type Router struct{}

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) Route(conn net.Conn) (ProtocolHandler, net.Conn, error) {
	bufferedReader := bufio.NewReader(conn)

	peekBytes, err := bufferedReader.Peek(1)
	if err != nil {
		return nil, conn, fmt.Errorf("failed to peek connection: %w", err)
	}

	if peekBytes[0] == 0x05 {
		return NewSocks5Handler(), &PeekedConn{Conn: conn, reader: bufferedReader}, nil
	}

	peekBytes, err = bufferedReader.Peek(10)
	if err != nil && err != io.EOF {
		return nil, conn, fmt.Errorf("failed to peek connection: %w", err)
	}
	if bytes.HasPrefix(peekBytes, []byte("CONNECT")) {
		return NewConnectHandler(), &PeekedConn{Conn: conn, reader: bufferedReader}, nil
	}

	return nil, conn, fmt.Errorf("unsupported protocol")
}

type PeekedConn struct {
	net.Conn
	reader *bufio.Reader
}

func (p *PeekedConn) Read(b []byte) (int, error) {
	return p.reader.Read(b)
}
