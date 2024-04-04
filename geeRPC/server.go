// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package geeRPC

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"seven-day-web-framework/geeRPC/codec"
	"seven-day-web-framework/geeRPC/request"
	"seven-day-web-framework/geeRPC/service"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int        // MagicNumber marks this's a geerpc request
	CodecType      codec.Type // client may choose different Codec to encode body
	ConnectTimeout time.Duration
	HandleTime     time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
	HandleTime:     time.Second * 10,
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

// Server represents an RPC Server.
type Server struct {
	serviceMap sync.Map
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{}
}

func (s *Server) Register(rcvr interface{}) error {
	newService := service.New(rcvr)
	if _, dup := s.serviceMap.LoadOrStore(newService.Name(), newService); dup {
		return errors.New("rpc: service already defined: " + newService.Name())
	}
	return nil
}

func (s *Server) FindService(serviceMethod string) (svc *service.Service, mtype *service.MethodType, err error) {
	dot := strings.LastIndex(serviceMethod, ".")
	if dot < 0 {
		err = errors.New("rpc server: service/method request ill-formed: " + serviceMethod)
		return
	}
	serviceName, methodName := serviceMethod[:dot], serviceMethod[dot+1:]
	svci, ok := s.serviceMap.Load(serviceName)
	if !ok {
		err = errors.New("rpc server: can't find service " + serviceName)
		return
	}
	svc = svci.(*service.Service)
	mtype, _ = svc.GetMethod(methodName)
	if mtype == nil {
		err = errors.New("rpc server: can't find method " + methodName)
	}
	return
}

func (s *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go s.ServeConn(conn)
	}
}

func (s *Server) ServeConn(conn io.ReadWriteCloser) {
	var opt Option
	defer func() { _ = conn.Close() }()

	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	s.serveCodec(f(conn))
}

func (s *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)
	for {
		req, err := s.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}
			req.H.Error = err.Error()
			s.sendResponse(cc, req.H, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go s.handleRequest(cc, req, sending, wg, time.Second)
	}
	wg.Wait()
	_ = cc.Close()
}

func (s *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && !errors.Is(err, io.ErrUnexpectedEOF) {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (s *Server) readRequest(cc codec.Codec) (*request.Request, error) {
	h, err := s.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request.Request{H: h}
	req.Service, req.MethodType, err = s.FindService(h.ServiceMethod)
	if err != nil {
		return req, err
	}
	req.Argv = req.MethodType.NewArgv()
	req.Replyv = req.MethodType.NewReplyv()

	argvi := req.Argv.Interface()
	if req.Argv.Type().Kind() != reflect.Ptr {
		argvi = req.Argv.Addr().Interface()
	}
	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read body err:", err)
		return req, err
	}
	return req, nil
}

func (s *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

//func (s *Server) handleRequest(cc codec.Codec, req *request.Request, sending *sync.Mutex, wg *sync.WaitGroup) {
//	log.Println("[ handleRequest ]", req.H, req.Argv)
//	err := req.Service.Call(req.MethodType, req.Argv, req.Replyv)
//	if err != nil {
//		req.H.Error = err.Error()
//		s.sendResponse(cc, req.H, invalidRequest, sending)
//		return
//	}
//	s.sendResponse(cc, req.H, req.Replyv.Interface(), sending)
//	wg.Done()
//}

func (s *Server) handleRequest(cc codec.Codec, req *request.Request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})
	go func() {
		err := req.Service.Call(req.MethodType, req.Argv, req.Replyv)
		called <- struct{}{}
		if err != nil {
			req.H.Error = err.Error()
			s.sendResponse(cc, req.H, invalidRequest, sending)
			sent <- struct{}{}
			return
		}
		s.sendResponse(cc, req.H, req.Replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.H.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		s.sendResponse(cc, req.H, invalidRequest, sending)
	case <-called:
		<-sent
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

func Register(rcvr interface{}) error {
	return DefaultServer.Register(rcvr)
}
