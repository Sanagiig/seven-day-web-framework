package geerpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"seven-day-web-framework/geeRPC2/codec"
	"strings"
	"sync"
	"time"
)

const MagicNumber = 0x3bef5c

type Option struct {
	MagicNumber    int           // MagicNumber marks this's a geerpc request
	CodecType      codec.Type    // client may choose different Codec to encode body
	ConnectTimeout time.Duration // 0 means no limit
	HandleTimeout  time.Duration
}

var DefaultOption = &Option{
	MagicNumber:    MagicNumber,
	CodecType:      codec.GobType,
	ConnectTimeout: time.Second * 10,
}

// request stores all information of a call
type request struct {
	h            *codec.Header // header of request
	svc          *service
	mtype        *methodType
	argv, replyv reflect.Value // argv and replyv of request
}

// Server represents an RPC Server.
type Server struct {
	serviceMap       sync.Map
	HandleReqTimeout time.Duration
}

// NewServer returns a new Server.
func NewServer() *Server {
	return &Server{HandleReqTimeout: time.Second * 3}
}

// DefaultServer is the default instance of *Server.
var DefaultServer = NewServer()

func (server *Server) Register(rscv any) error {
	s := newService(rscv)
	if _, isExist := server.serviceMap.LoadOrStore(s.name, s); isExist {
		return errors.New("rpc: service already defined: " + s.name)
	}
	return nil
}

// Register publishes the receiver's methods in the DefaultServer.
func Register(rcvr interface{}) error { return DefaultServer.Register(rcvr) }

func (server *Server) findServer(serviceMethodName string) (svc *service, mtype *methodType, err error) {
	dotIdx := strings.LastIndex(serviceMethodName, ".")
	if dotIdx == -1 {
		err = fmt.Errorf("rpc: serviceMethodName [%s] not had [.]", serviceMethodName)
		return
	}

	svcName := serviceMethodName[:dotIdx]
	methodName := serviceMethodName[dotIdx+1:]
	svcv, ok := server.serviceMap.Load(svcName)
	if !ok {
		return nil, nil, fmt.Errorf("service not exist [%s]", serviceMethodName)
	}

	svc = svcv.(*service)
	mtype = svc.method[methodName]
	if mtype == nil {
		err = fmt.Errorf("methodType [%s] not not exist in [%s]", methodName, svcName)
		return
	}

	return
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func (server *Server) Accept(lis net.Listener) {
	for {
		conn, err := lis.Accept()
		if err != nil {
			log.Println("rpc server: accept error:", err)
			return
		}
		go server.ServeConn(conn)
	}
}

// Accept accepts connections on the listener and serves requests
// for each incoming connection.
func Accept(lis net.Listener) { DefaultServer.Accept(lis) }

// ServeConn runs the server on a single connection.
// ServeConn blocks, serving the connection until the client hangs up.
func (server *Server) ServeConn(conn io.ReadWriteCloser) {
	defer func() { _ = conn.Close() }()
	var opt Option
	if err := json.NewDecoder(conn).Decode(&opt); err != nil {
		log.Println("rpc server: options error: ", err)
		return
	}
	if opt.MagicNumber != MagicNumber {
		log.Printf("rpc server: invalid magic number %x", opt.MagicNumber)
		return
	}
	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Printf("rpc server: invalid codec type %s", opt.CodecType)
		return
	}
	server.serveCodec(f(conn))
}

// invalidRequest is a placeholder for response argv when error occurs
var invalidRequest = struct{}{}

func (server *Server) serveCodec(cc codec.Codec) {
	sending := new(sync.Mutex) // make sure to send a complete response
	wg := new(sync.WaitGroup)  // wait until all request are handled
	for {
		req, err := server.readRequest(cc)
		if err != nil {
			if req == nil {
				break // it's not possible to recover, so close the connection
			}
			req.h.Error = err.Error()
			server.sendResponse(cc, req.h, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go server.handleRequest(cc, req, sending, wg, server.HandleReqTimeout)
	}
	wg.Wait()
	_ = cc.Close()
}

func (server *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var h codec.Header
	if err := cc.ReadHeader(&h); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server: read header error:", err)
		}
		return nil, err
	}
	return &h, nil
}

func (server *Server) readRequest(cc codec.Codec) (*request, error) {
	h, err := server.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}
	req := &request{h: h}

	req.svc, req.mtype, err = server.findServer(h.ServiceMethod)
	if err != nil {
		err = fmt.Errorf("find server err: %w", err)
		return nil, err
	}

	req.argv, req.replyv = req.mtype.newArgv(), req.mtype.newReplyv()
	argvi := req.argv.Interface()
	if req.argv.Type().Kind() != reflect.Pointer {
		argvi = req.argv.Addr().Interface()
	}

	if err = cc.ReadBody(argvi); err != nil {
		log.Println("rpc server: read argv err:", err)
	}

	return req, nil
}

func (server *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc server: write response error:", err)
	}
}

func (server *Server) handleRequest(cc codec.Codec, req *request, sending *sync.Mutex, wg *sync.WaitGroup, timeout time.Duration) {
	defer wg.Done()
	called := make(chan struct{})
	sent := make(chan struct{})

	go func() {
		err := req.svc.call(req.mtype, req.argv, req.replyv)
		called <- struct{}{}
		if err != nil {
			server.sendResponse(cc, req.h, invalidRequest, sending)
			sent <- struct{}{}
		}

		server.sendResponse(cc, req.h, req.replyv.Interface(), sending)
		sent <- struct{}{}
	}()

	if timeout == 0 {
		<-called
		<-sent
		return
	}
	select {
	case <-time.After(timeout):
		req.h.Error = fmt.Sprintf("rpc server: request handle timeout: expect within %s", timeout)
		server.sendResponse(cc, req.h, invalidRequest, sending)
	case <-called:
		<-sent
	}

}
