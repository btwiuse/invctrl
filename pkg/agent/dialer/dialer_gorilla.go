// +build gorilla

package dialer

import (
	"net"
	"path"

	"github.com/btwiuse/conntroll/pkg/wrap"
	"github.com/gorilla/websocket"
)

func (d *dialr) Dial(p string) (conn net.Conn, err error) {
	var (
		c = d.c
		u string
	)

	switch c.GetScheme() {
	case "http":
		u = "ws://" + path.Join(c.GetAddr(), p)
	case "https":
		u = "wss://" + path.Join(c.GetAddr(), p)
	}

	wsconn, _, err := websocket.DefaultDialer.Dial(u, nil)

	if err != nil {
		return nil, err
	}

	return wrap.NetConn(wsconn), nil
}
