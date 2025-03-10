package server

import (
	"net"

	"k0s.io/pkg/agent/config"
)

func StartPingServer(c *config.Config) chan net.Conn {
	pingListener := NewChannelListener()
	go pingServe(pingListener)
	return pingListener.Conns
}

func pingServe(l net.Listener) {
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}

		go func() {
			defer c.Close()
			// nop
		}()
	}
}
