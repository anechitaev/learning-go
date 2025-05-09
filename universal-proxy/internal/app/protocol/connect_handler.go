package protocol

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type ConnectHandler struct{}

func NewConnectHandler() *ConnectHandler {
	return &ConnectHandler{}
}

func (h *ConnectHandler) Handle(clientConn net.Conn) error {
	var wg sync.WaitGroup
	log.Printf("Accepted connection from %s", clientConn.RemoteAddr())

	reader := bufio.NewReader(clientConn)
	host, ok := tryToFindHost(reader)
	if !ok {
		fmt.Fprintf(clientConn, "%s\r\n\r\n", CONNECTION_FORBIDDEN)
		log.Printf("Refused non-CONNECT request")
		return fmt.Errorf("not a CONNECT request")
	}

	serverConn, err := net.Dial("tcp", host)
	if err != nil {
		log.Printf("Accept error: %v", err)
		return err
	}
	log.Printf("Establish connection to %s", host)

	defer clientConn.Close()
	defer serverConn.Close()

	fmt.Fprintf(clientConn, "%s\r\n\r\n", CONNECTION_ESTABLISHED)
	log.Printf("Start copying")

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

// Вспомогательные функции сюда тоже пока можно скопировать:
func tryToFindHost(reader *bufio.Reader) (string, bool) {
	line, err := reader.ReadString('\n')
	if err != nil {
		log.Printf("Error while reading input channel")
		return "", false
	}
	splittedLine := strings.SplitN(line, " ", 3)
	method := splittedLine[0]
	if method != "CONNECT" {
		return "", false
	}
	maybeHost := splittedLine[1]
	log.Printf("Found host: %s", maybeHost)
	return maybeHost, true
}

type copyCall func()

func deferCopyCall(wg *sync.WaitGroup, cc copyCall) {
	defer wg.Done()
	cc()
}

var CONNECTION_ESTABLISHED = "HTTP/1.1 200 Connection Established"
var CONNECTION_FORBIDDEN = "HTTP/1.1 403 Forbidden"
