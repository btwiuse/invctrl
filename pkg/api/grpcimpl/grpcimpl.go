package grpcimpl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"path/filepath"
	"strings"

	"github.com/btwiuse/wetty/pkg/msg"
	"k0s.io/conntroll/pkg/agent"
	"k0s.io/conntroll/pkg/api"

	"go.uber.org/zap"
)

type Session struct {
	ReadOnly       bool
	TtyFactory     agent.TtyFactory
	FileServer     http.Handler
	MetricsHandler http.Handler
	// client id/index, to distinguish logs of different commands
}

func (session *Session) Chunker(req *api.ChunkRequest, chunkerServer api.Session_ChunkerServer) (err error) {
	// req.Path
	// req.Request
	// log.Println("Chunker called with", req)

	// no such file => fileserver
	// isdir => fileserver
	// openerror => fileserver
	// other => chunker

	var (
		path = filepath.Clean(req.Path)

		statfail bool
		openfail bool
		isdir    bool
		issmall  bool

		filename string
		filesize int64

		reader io.Reader
	)

	info, staterr := os.Stat(path)
	if staterr == nil {
		if !info.IsDir() {
			filesize = info.Size()
			if filesize > 4*(1<<20) { // 4M
				f, openerr := os.Open(path)
				if openerr == nil {
					filename = filepath.Base(f.Name())
					defer f.Close()
					header := fmt.Sprintf(ResponseHeaderTemplate, filename, filesize)
					h := strings.NewReader(header)
					reader = io.MultiReader(h, f)
				} else {
					openfail = true
				}
			} else {
				issmall = true
			}
		} else {
			isdir = true
		}
	} else {
		statfail = true
	}

	switch {
	case path == "metrics":
		w := httptest.NewRecorder()
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(req.Request)))
		if err != nil {
			return err
		}
		session.MetricsHandler.ServeHTTP(w, r)
		resp, err := httputil.DumpResponse(w.Result(), true)
		if err != nil {
			return err
		}
		// fmt.Printf("%s", resp)
		reader = bytes.NewReader(resp)
	case statfail || openfail || isdir || issmall:
		w := httptest.NewRecorder()
		r, err := http.ReadRequest(bufio.NewReader(bytes.NewBuffer(req.Request)))
		if err != nil {
			return err
		}
		session.FileServer.ServeHTTP(w, r)
		resp, err := httputil.DumpResponse(w.Result(), true)
		if err != nil {
			return err
		}
		// fmt.Printf("%s", resp)
		reader = bytes.NewReader(resp)
	default:
		log.Println("sending large chunked file:", filename)
	}

	ns := []int{}
	defer func() {
		if err != nil {
			log.Println(ns, err)
		}
	}()

	buf := make([]byte, 64*1024)
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		ns = append(ns, n)
		if err != nil {
			return err
		}
		chunk := &api.Chunk{
			Chunk: buf[:n],
		}
		err = chunkerServer.Send(chunk)
		if err != nil {
			return err
		}
	}
	return nil
}

func (session *Session) Send(sendServer api.Session_SendServer) error {
	recorder, _ := zap.NewProduction()
	defer recorder.Sync()
	tty, err := session.TtyFactory.MakeTty()
	if err != nil {
		return err
	}

	// send
	go func() {
		buf := make([]byte, 1<<12-1)
		if err != nil {
			return // err
		}
		for {
			n, err := tty.Read(buf)
			if err == io.EOF {
				return // nil
			}
			if err != nil {
				return // err
			}
			var (
				msgType = msg.Type_SESSION_OUTPUT
				msgBody = buf[:n]
				req     = &api.Message{
					Type: msgType,
					Body: msgBody,
				}
			)
			err = sendServer.Send(req)
			if err != nil {
				return // err
			}
			// log.Println(req.Type, fmt.Sprintf("%q", string(req.Body)))
			/* this causes infinite log loop, be careful
			recorder.Info("send",
				zap.String("type", req.Type.String()),
				zap.String("content", string(req.Body)),
			)
			*/
		}
		return // nil
	}()

	// recv
	for {
		resp, err := sendServer.Recv()
		if err != nil {
			return nil
		}
		// log.Println(resp.Type, fmt.Sprintf("%q", string(resp.Body)))
		recorder.Info("recv",
			zap.String("type", resp.Type.String()),
			zap.String("content", string(resp.Body)),
		)
		switch resp.Type {
		case msg.Type_CLIENT_INPUT:
			if session.ReadOnly {
				break
			}
			_, err = tty.Write(resp.Body)
			if err != nil {
				log.Println("error writing to tty:", err)
				return err
			}
		case msg.Type_SESSION_RESIZE:
			type Winsize struct {
				Rows int
				Cols int
			}
			sz := &Winsize{}
			err = json.Unmarshal(resp.Body, sz)
			if err != nil {
				return err
			}
			err = tty.Resize(sz.Rows, sz.Cols)
			if err != nil {
				return err
			}
		case msg.Type_SESSION_CLOSE:
			return tty.Close()
		}
	}

	return nil
}

var ResponseHeaderTemplate = strings.Join([]string{
	"HTTP/1.1 200 OK",
	"Accept-Ranges: bytes",
	"Content-Type: application/octet-stream",
	"Content-Disposition: attachment; filename=%q",
	"Last-Modified: Sun, 29 Sep 2019 03:58:56 GMT",
	"Content-Length: %d",
	"\r\n"}, "\r\n")