package protocol

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Socks5Handler struct{}

// Handle implements ProtocolHandler.
func (s *Socks5Handler) Handle(clientConn net.Conn) error {
	reader := bufio.NewReader(clientConn)

	// 1. Читаем версию
	ver, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	if ver != 0x05 {
		return fmt.Errorf("unsupported socks version: %d", ver)
	}

	// 2. Читаем количество методов
	nMethods, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read nmethods: %w", err)
	}

	// 3. Читаем список методов (можно пропустить сами значения)
	methods := make([]byte, nMethods)
	if _, err := io.ReadFull(reader, methods); err != nil {
		return fmt.Errorf("failed to read methods: %w", err)
	}

	// 4. Отвечаем клиенту: выбрали метод 0x00 (без аутентификации)
	_, err = clientConn.Write([]byte{0x05, 0x00})
	if err != nil {
		return fmt.Errorf("failed to write handshake response: %w", err)
	}

	// 👉 После этого handshake закончен, можно переходить к чтению команды CONNECT
	ver, err = reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read version: %w", err)
	}
	if ver != 0x05 {
		return fmt.Errorf("unsupported socks version: %d", ver)
	}

	cmd, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read cmd: %w", err)
	}
	if cmd != 0x01 {
		return fmt.Errorf("unsupported command: %d", cmd)
	}

	rsv, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read rsv: %w", err)
	}
	if rsv != 0x00 {
		return fmt.Errorf("invalid reserved field: %d", rsv)
	}
	// 4. Читаем тип адреса
	atyp, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("failed to read type: %w", err)
	}
	var addr string

	switch atyp {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(reader, ip); err != nil {
			return fmt.Errorf("failed to read ipv4 address: %w", err)
		}
		addr = net.IP(ip).String()

	case 0x03: // Domain Name
		domainLen, err := reader.ReadByte()
		if err != nil {
			return fmt.Errorf("failed to read domain length: %w", err)
		}
		domain := make([]byte, domainLen)
		if _, err := io.ReadFull(reader, domain); err != nil {
			return fmt.Errorf("failed to read domain: %w", err)
		}
		addr = string(domain)

	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(reader, ip); err != nil {
			return fmt.Errorf("failed to read ipv6 address: %w", err)
		}
		addr = net.IP(ip).String()

	default:
		return fmt.Errorf("unsupported address type: %d", atyp)
	}

	portBytes := make([]byte, 2)
	if _, err := io.ReadFull(reader, portBytes); err != nil {
		return fmt.Errorf("failed to read port: %w", err)
	}
	port := int(portBytes[0])<<8 | int(portBytes[1])

	// Логируем результат
	log.Printf("Client requested connection to %s:%d", addr, port)

	serverConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return fmt.Errorf("failed to connect to target server: %w", err)
	}

	log.Printf("Establish connection to %s:%d", addr, port)

	_, err = clientConn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		return fmt.Errorf("failed to write connect response: %w", err)
	}
	var wg sync.WaitGroup

	wg.Add(2)
	go deferCopyCall(&wg, func() {
		io.Copy(serverConn, clientConn)
	})
	go deferCopyCall(&wg, func() {
		io.Copy(clientConn, serverConn)
	})
	wg.Wait()

	return nil
}

func NewSocks5Handler() *Socks5Handler {
	return &Socks5Handler{}
}
