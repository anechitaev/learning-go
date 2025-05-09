package app

import (
	"context"
	"log"
	"net"
	"universal-proxy/internal/app/protocol"
	"universal-proxy/internal/config"
)

type ProxyApp struct {
	cfg config.Config
}

func NewProxyApp(cfg config.Config) *ProxyApp {
	return &ProxyApp{cfg: cfg}
}

func (p *ProxyApp) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", p.cfg.ListenAddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Printf("Proxy listening on %s", p.cfg.ListenAddr)

	go func() {
		<-ctx.Done()
		log.Println("Context canceled, shutting down listener...")
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // Штатное завершение
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}
		go p.handleConnection(conn)
	}
}

func (p *ProxyApp) handleConnection(clientConn net.Conn) {
	router := protocol.NewRouter()

	handler, routedConn, err := router.Route(clientConn)
	if err != nil {
		log.Printf("Unsupported protocol from %s: %v", clientConn.RemoteAddr(), err)
		clientConn.Close()
		return
	}

	if err := handler.Handle(routedConn); err != nil {
		log.Printf("Connection handler error: %v", err)
	}
}
