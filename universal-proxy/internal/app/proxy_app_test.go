package app

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"testing"
	"time"
)

var CONNECTION_ESTABLISHED = "HTTP/1.1 200 Connection Established"
var CONNECTION_FORBIDDEN = "HTTP/1.1 403 Forbidden"

func TestProxy_CONNECT_Success(t *testing.T) {
	testCaseRaw(
		[]byte("CONNECT www.google.com:443 HTTP/1.1\r\nHost: www.google.com:443\r\n\r\n"), []byte(CONNECTION_ESTABLISHED), t,
	)

}

func TestProxy_CONNECT_Forbiden(t *testing.T) {
	testCaseRaw(
		[]byte("GET www.google.com:443 HTTP/1.1\r\nHost: www.google.com:443\r\n\r\n"), []byte(CONNECTION_FORBIDDEN), t,
	)
}

func TestProxy_SOCKS5_ConnectRequest(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				p := &ProxyApp{}
				p.handleConnection(c)
			}(conn)
		}
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer clientConn.Close()

	reader := bufio.NewReader(clientConn)

	// 1. Отправляем Hello
	hello := []byte{0x05, 0x01, 0x00}
	_, err = clientConn.Write(hello)
	if err != nil {
		t.Fatalf("Failed to write hello: %v", err)
	}

	// 2. Читаем ответ на Hello
	resp := make([]byte, 2)
	if _, err := io.ReadFull(reader, resp); err != nil {
		t.Fatalf("Failed to read hello response: %v", err)
	}
	if !bytes.Equal(resp, []byte{0x05, 0x00}) {
		t.Fatalf("Unexpected hello response: %v", resp)
	}

	// 3. Отправляем Request (на какой-нибудь несуществующий адрес для теста)
	domain := "google.com"
	req := []byte{
		0x05, 0x01, 0x00, 0x03, // VER, CMD=CONNECT, RSV, ATYP=DOMAIN
		byte(len(domain)), // длина домена
	}
	req = append(req, []byte(domain)...) // сам домен
	req = append(req, 0x00, 0x50)        // порт 80

	_, err = clientConn.Write(req)
	if err != nil {
		t.Fatalf("Failed to write connect request: %v", err)
	}

	// 4. Читаем ответ на CONNECT
	connResp := make([]byte, 10)
	if _, err := io.ReadFull(reader, connResp); err != nil {
		t.Fatalf("Failed to read connect response: %v", err)
	}
	expected := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	if !bytes.Equal(connResp, expected) {
		t.Fatalf("Unexpected connect response: got %v, expected %v", connResp, expected)
	}
}

func testCaseRaw(requestBytes []byte, expectedResponseBytes []byte, t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to start listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				p := &ProxyApp{}
				p.handleConnection(c)
			}(conn)
		}
	}()

	clientConn, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to proxy: %v", err)
	}
	defer clientConn.Close()

	// Отправляем запрос
	_, err = clientConn.Write(requestBytes)
	if err != nil {
		t.Fatalf("Failed to write to proxy: %v", err)
	}

	clientConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	// Читаем ответ
	response := make([]byte, len(expectedResponseBytes))
	_, err = io.ReadFull(clientConn, response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	// Сравниваем байты
	if !bytes.Equal(response, expectedResponseBytes) {
		t.Errorf("Unexpected response. Got: %v, Expected: %v", response, expectedResponseBytes)
	}
}
