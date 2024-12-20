package server

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/btwiuse/wsconn"
	"github.com/gorilla/mux"
	"k0s.io"
	"k0s.io/pkg/api"
	"k0s.io/pkg/hub"
	"nhooyr.io/websocket"
)

func terminalV2Relay(ag hub.Agent) http.HandlerFunc {
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

func terminalRelay(ag hub.Agent) http.HandlerFunc {
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

func fsRelay(ag hub.Agent) http.HandlerFunc {
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

		conn, err := wsconn.Hijack(w)
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

func versionRelay(ag hub.Agent) http.HandlerFunc {
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

		conn, err := wsconn.Hijack(w)
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

func dohRelay(ag hub.Agent) http.HandlerFunc {
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

		conn, err := wsconn.Hijack(w)
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

func jsonlRelay(ag hub.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			log.Println(err)
			return
		}
		wsconn := websocket.NetConn(context.Background(), wsc, websocket.MessageBinary)
		defer wsconn.Close()

		conn := ag.NewTunnel(api.Jsonl)
		defer conn.Close()

		go io.Copy(conn, wsconn)
		io.Copy(wsconn, conn)
	}
}

func xpraRelay(ag hub.Agent) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wsc, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
			Subprotocols:       []string{"binary"},
			CompressionMode:    websocket.CompressionContextTakeover,
		})
		if err != nil {
			log.Println(err)
			return
		}
		wsc.SetReadLimit(k0s.MAX_WS_MESSAGE)

		wsconn := websocket.NetConn(context.Background(), wsc, websocket.MessageBinary)
		defer wsconn.Close()

		conn := ag.NewTunnel(api.Xpra)
		defer conn.Close()

		b := make([]byte, k0s.MAX_WS_MESSAGE)
		go io.CopyBuffer(conn, wsconn, b)

		// b := make([]byte, 160*1024)
		io.Copy(wsconn, conn)
	}
}

func envRelay(ag hub.Agent) http.HandlerFunc {
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

		conn, err := wsconn.Hijack(w)
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

func socks5Relay(ag hub.Agent) http.HandlerFunc {
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

func redirRelay(ag hub.Agent) http.HandlerFunc {
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
