package client

import (
	"fmt"
	"net"
	"seven-day-web-framework/geeRPC"
	"time"
)

type ClientResult struct {
	Client *Client
	Err    error
}

type NewClientFunc func(conn net.Conn, opt *geeRPC.Option) (client *Client, err error)

func DialTimeout(f NewClientFunc, netwoork, address string, opts ...*geeRPC.Option) (client *Client, err error) {
	opt, err := parseOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout(netwoork, address, opt.ConnectTimeout)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = conn.Close()
		}
	}()
	ch := make(chan ClientResult)
	go func() {
		client, err = f(conn, opt)
		ch <- ClientResult{Client: client, Err: err}
	}()
	if opt.ConnectTimeout == 0 {
		result := <-ch
		return result.Client, result.Err
	}
	select {
	case <-time.After(opt.ConnectTimeout):
		return nil, fmt.Errorf("rpc client: connect timeout: expect within %s", opt.ConnectTimeout)
	case result := <-ch:
		return result.Client, result.Err
	}
}
