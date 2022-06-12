package agent

import (
	"log"
	"net"

	types "k0s.io/pkg/agent"
	"k0s.io/pkg/agent/tty"
	"k0s.io/pkg/api"
	"k0s.io/pkg/asciitransport"
)

func init() { Tunnels[api.Terminal] = StartTerminalServer }

func StartTerminalServer(c types.Config) chan net.Conn {
	var (
		cmd              []string = c.GetCmd()
		ro               bool     = c.GetReadOnly()
		fac                       = tty.NewFactory(cmd)
		terminalListener          = NewLys()
	)
	_ = ro
	go serveTerminal(terminalListener, fac)
	return terminalListener.Conns
}

func serveTerminal(ln net.Listener, fac types.TtyFactory) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go func() {
			term, err := fac.MakeTty()
			if err != nil {
				log.Println(err)
				return
			}

			opts := []asciitransport.Opt{
				asciitransport.WithReader(term),
				asciitransport.WithWriter(term),
			}
			server := asciitransport.Server(conn, opts...)
			// send
			// case output:

			// recv
			go func() {
				for {
					var (
						re   = <-server.ResizeEvent()
						rows = int(re.Height)
						cols = int(re.Width)
					)
					err := term.Resize(rows, cols)
					if err != nil {
						log.Println(err)
						break
					}
				}
				server.Close()
			}()
		}()
	}
}
