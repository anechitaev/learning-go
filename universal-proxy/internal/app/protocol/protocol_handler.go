package protocol

import "net"

type ProtocolHandler interface {
	Handle(clientConn net.Conn) error
}
