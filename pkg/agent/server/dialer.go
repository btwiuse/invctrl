package server

import (
	"net"
	"net/url"

	"github.com/btwiuse/wsdial"
	"k0s.io/pkg/agent/config"
)

type dialer struct {
	c *config.Config
}

func (d *dialer) Dial(p string, q string) (conn net.Conn, err error) {
	u := &url.URL{
		Scheme:   d.c.GetSchemeWS(),
		Host:     d.c.GetAddr(),
		Path:     p,
		RawQuery: q,
	}

	return wsdial.Dial(u)
}
