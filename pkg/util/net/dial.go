package net

import (
	"context"
	"net"
	//"net/url"
	"time"

	libdial "github.com/fatedier/golib/net/dial"
	"github.com/gorilla/websocket"
	//"golang.org/x/net/websocket"
)

func DialHookCustomTLSHeadByte(enableTLS bool, disableCustomTLSHeadByte bool) libdial.AfterHookFunc {
	return func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error) {
		if enableTLS && !disableCustomTLSHeadByte {
			_, err := c.Write([]byte{byte(FRPTLSHeadByte)})
			if err != nil {
				return nil, nil, err
			}
		}
		return ctx, c, nil
	}
}

func DialHookWebsocket(protocol string, host string, websocketPath string) libdial.AfterHookFunc {
	return func(ctx context.Context, c net.Conn, addr string) (context.Context, net.Conn, error) {
		if protocol != "wss" {
			protocol = "ws"
		}
		if host == "" {
			host = addr
		}
		addr = protocol + "://" + host + websocketPath

		dialer := &websocket.Dialer{
			NetDial: func(network, addr string) (net.Conn, error) {
				return c, nil
			},
			ReadBufferSize:   4 * 1024,
			WriteBufferSize:  4 * 1024,
			HandshakeTimeout: time.Second * 8,
		}
		wsConn, _, err := dialer.DialContext(ctx, addr, nil)
		if err != nil {
			return nil, nil, err
		}
		conn := &websocketConn{conn: wsConn}
		return ctx, conn, nil

		//uri, err := url.Parse(addr)
		//if err != nil {
		//	return nil, nil, err
		//}
		//
		//origin := "http://" + uri.Host
		//cfg, err := websocket.NewConfig(addr, origin)
		//if err != nil {
		//	return nil, nil, err
		//}
		//
		//conn, err := websocket.NewClient(cfg, c)
		//if err != nil {
		//	return nil, nil, err
		//}
		//return ctx, conn, nil
	}
}
