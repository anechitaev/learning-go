package main

import (
	"context"
	"log"

	"universal-proxy/internal/app"
	"universal-proxy/internal/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Load()
	proxy := app.NewProxyApp(cfg)

	go func() {
		// Здесь можно перехватывать Ctrl+C через signal.Notify
	}()

	if err := proxy.Start(ctx); err != nil {
		log.Fatalf("Proxy stopped with error: %v", err)
	}
}
