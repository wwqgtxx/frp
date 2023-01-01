package net

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	//"golang.org/x/net/websocket"
)

var ErrWebsocketListenerClosed = errors.New("websocket listener closed")

type WebsocketListener struct {
	ln       net.Listener
	acceptCh chan net.Conn

	server *http.Server

	upgrader websocket.Upgrader
}

// NewWebsocketListener to handle websocket connections
// ln: tcp listener for websocket connections
func NewWebsocketListener(ln net.Listener, websocketPath string) (wl *WebsocketListener) {
	wl = &WebsocketListener{
		acceptCh: make(chan net.Conn),
	}

	muxer := http.NewServeMux()
	muxer.HandleFunc(websocketPath, func(w http.ResponseWriter, r *http.Request) {
		wsConn, err := wl.upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		c := &websocketConn{conn: wsConn}
		notifyCh := make(chan struct{})
		conn := WrapCloseNotifyConn(c, func() {
			close(notifyCh)
		})
		wl.acceptCh <- conn
		<-notifyCh
	})
	//muxer.Handle(websocketPath, websocket.Handler(func(c *websocket.Conn) {
	//	notifyCh := make(chan struct{})
	//	conn := WrapCloseNotifyConn(c, func() {
	//		close(notifyCh)
	//	})
	//	wl.acceptCh <- conn
	//	<-notifyCh
	//}))

	wl.server = &http.Server{
		Addr:    ln.Addr().String(),
		Handler: muxer,
	}

	go func() {
		_ = wl.server.Serve(ln)
	}()
	return
}

func ListenWebsocket(bindAddr string, bindPort int) (*WebsocketListener, error) {
	tcpLn, err := net.Listen("tcp", net.JoinHostPort(bindAddr, strconv.Itoa(bindPort)))
	if err != nil {
		return nil, err
	}
	l := NewWebsocketListener(tcpLn, "/~!frp")
	return l, nil
}

func (p *WebsocketListener) Accept() (net.Conn, error) {
	c, ok := <-p.acceptCh
	if !ok {
		return nil, ErrWebsocketListenerClosed
	}
	return c, nil
}

func (p *WebsocketListener) Close() error {
	return p.server.Close()
}

func (p *WebsocketListener) Addr() net.Addr {
	return p.ln.Addr()
}

// modify from clash's transport/vmess/websocket.go
type websocketConn struct {
	conn   *websocket.Conn
	reader io.Reader

	// https://godoc.org/github.com/gorilla/websocket#hdr-Concurrency
	rMux sync.Mutex
	wMux sync.Mutex
}

// Read implements net.Conn.Read()
func (wsc *websocketConn) Read(b []byte) (int, error) {
	wsc.rMux.Lock()
	defer wsc.rMux.Unlock()
	for {
		reader, err := wsc.getReader()
		if err != nil {
			return 0, err
		}

		nBytes, err := reader.Read(b)
		if err == io.EOF {
			wsc.reader = nil
			continue
		}
		return nBytes, err
	}
}

// Write implements io.Writer.
func (wsc *websocketConn) Write(b []byte) (int, error) {
	wsc.wMux.Lock()
	defer wsc.wMux.Unlock()
	if err := wsc.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return 0, err
	}
	return len(b), nil
}

func (wsc *websocketConn) Close() error {
	var errors []string
	if err := wsc.conn.WriteControl(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""), time.Now().Add(time.Second*5)); err != nil {
		errors = append(errors, err.Error())
	}
	if err := wsc.conn.Close(); err != nil {
		errors = append(errors, err.Error())
	}
	if len(errors) > 0 {
		return fmt.Errorf("failed to close connection: %s", strings.Join(errors, ","))
	}
	return nil
}

func (wsc *websocketConn) getReader() (io.Reader, error) {
	if wsc.reader != nil {
		return wsc.reader, nil
	}

	_, reader, err := wsc.conn.NextReader()
	if err != nil {
		return nil, err
	}
	wsc.reader = reader
	return reader, nil
}

func (wsc *websocketConn) LocalAddr() net.Addr {
	return wsc.conn.LocalAddr()
}

func (wsc *websocketConn) RemoteAddr() net.Addr {
	return wsc.conn.RemoteAddr()
}

func (wsc *websocketConn) SetDeadline(t time.Time) error {
	if err := wsc.SetReadDeadline(t); err != nil {
		return err
	}
	return wsc.SetWriteDeadline(t)
}

func (wsc *websocketConn) SetReadDeadline(t time.Time) error {
	return wsc.conn.SetReadDeadline(t)
}

func (wsc *websocketConn) SetWriteDeadline(t time.Time) error {
	return wsc.conn.SetWriteDeadline(t)
}
