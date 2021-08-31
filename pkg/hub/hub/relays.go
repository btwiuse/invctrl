package hub

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/gorilla/mux"
	"k0s.io/pkg/api"
	types "k0s.io/pkg/hub"
	"k0s.io/pkg/wrap"
	"nhooyr.io/websocket"
)

func terminalV2Relay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			Subprotocols:       []string{"wetty"},
		})
		if err != nil {
			log.Println(err)
			return
		}
		wsconn := websocket.NetConn(context.Background(), wsc, websocket.MessageBinary)
		defer wsconn.Close()

		conn := ag.NewTunnel(api.TerminalV2)
		defer conn.Close()

		go io.Copy(conn, wsconn)
		io.Copy(wsconn, conn)
	}
}

func terminalRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			Subprotocols:       []string{"wetty"},
		})
		if err != nil {
			log.Println(err)
			return
		}
		wsconn := websocket.NetConn(context.Background(), wsc, websocket.MessageBinary)
		defer wsconn.Close()

		conn := ag.NewTunnel(api.Terminal)
		defer conn.Close()

		go io.Copy(conn, wsconn)
		io.Copy(wsconn, conn)
	}
}

func fsRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id+"/rootfs")
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		fsConn := ag.NewTunnel(api.FS)
		defer fsConn.Close()

		go func() {
			io.Copy(fsConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, fsConn)
	}
}

func versionRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id)
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		versionConn := ag.NewTunnel(api.Version)
		defer versionConn.Close()

		go func() {
			io.Copy(versionConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, versionConn)
	}
}

func dohRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id)
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		metricsConn := ag.NewTunnel(api.Doh)
		defer metricsConn.Close()

		go func() {
			io.Copy(metricsConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, metricsConn)
	}
}

func envRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id)
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		envConn := ag.NewTunnel(api.Env)
		defer envConn.Close()

		go func() {
			io.Copy(envConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, envConn)
	}
}

func k16sRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id)
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		metricsConn := ag.NewTunnel(api.K16s)
		defer metricsConn.Close()

		go func() {
			io.Copy(metricsConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, metricsConn)
	}
}

func metricsRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			vars = mux.Vars(r)
			id   = vars["id"]
			path = strings.TrimPrefix(r.RequestURI, "/api/agent/"+id)
		)
		r.RequestURI = path

		reqbuf, err := httputil.DumpRequest(r, true)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := wrap.Hijack(w)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()

		metricsConn := ag.NewTunnel(api.Metrics)
		defer metricsConn.Close()

		go func() {
			io.Copy(metricsConn, bytes.NewBuffer(reqbuf))
		}()
		io.Copy(conn, metricsConn)
	}
}

func socks5Relay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println(err)
			return
		}
		conn := websocket.NetConn(context.Background(), wsconn, websocket.MessageBinary)

		socks5Conn := ag.NewTunnel(api.Socks5)
		defer socks5Conn.Close()

		go func() {
			_, err := io.Copy(conn, socks5Conn)
			if err != nil {
				log.Println(err)
				return
			}
		}()

		_, err = io.Copy(socks5Conn, conn)
		if err != nil {
			log.Println(err)
			return
		}
	}
}

func redirRelay(ag types.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsconn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println(err)
			return
		}
		conn := websocket.NetConn(context.Background(), wsconn, websocket.MessageBinary)

		redirConn := ag.NewTunnel(api.Redir)
		defer redirConn.Close()

		go func() {
			_, err := io.Copy(conn, redirConn)
			if err != nil {
				log.Println(err)
				return
			}
		}()

		_, err = io.Copy(redirConn, conn)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
