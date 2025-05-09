package config

import "os"

type Config struct {
	ListenAddr string
}

func Load() Config {
	listenAddr := os.Getenv("PROXY_LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = ":1080" // default
	}
	return Config{
		ListenAddr: listenAddr,
	}
}
