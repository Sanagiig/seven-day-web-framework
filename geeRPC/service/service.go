package service

import (
	"go/ast"
	"log"
	"reflect"
	"sync/atomic"
)

type Service struct {
	name   string
	typ    reflect.Type
	rcvr   reflect.Value
	method map[string]*MethodType
}

func (s *Service) Name() string {
	return s.name
}

func (s *Service) registerMethods() {
	s.method = make(map[string]*MethodType)
	for i := 0; i < s.typ.NumMethod(); i++ {
		method := s.typ.Method(i)
		mType := method.Type
		if mType.NumIn() != 3 || mType.NumOut() != 1 {
			continue
		}
		if mType.Out(0) != reflect.TypeOf((*error)(nil)).Elem() {
			continue
		}
		argType, replyType := mType.In(1), mType.In(2)
		if !IsExportedOrBuiltinType(argType) || !IsExportedOrBuiltinType(replyType) {
			continue
		}
		s.method[method.Name] = &MethodType{
			method:    method,
			ArgType:   argType,
			ReplyType: replyType,
		}
		log.Printf("rpc server: register %s.%s\n", s.name, method.Name)
	}
}

func (s *Service) call(m *MethodType, argv, replyv reflect.Value) error {
	atomic.AddUint64(&m.numCalls, 1)

	f := m.method.Func
	returnValues := f.Call([]reflect.Value{s.rcvr, argv, replyv})
	if errInter := returnValues[0].Interface(); errInter != nil {
		return errInter.(error)
	}
	return nil
}

func (s *Service) Call(m *MethodType, argv, replyv reflect.Value) error {
	return s.call(m, argv, replyv)
}

func (s *Service) GetMethod(key string) (method *MethodType, ok bool) {
	method, ok = s.method[key]
	return
}

func New(rcvr interface{}) *Service {
	s := new(Service)
	s.rcvr = reflect.ValueOf(rcvr)
	s.name = reflect.Indirect(s.rcvr).Type().Name()
	s.typ = reflect.TypeOf(rcvr)
	if !ast.IsExported(s.name) {
		log.Fatalf("rpc server: %s is not a valid service name", s.name)
	}
	s.registerMethods()
	return s
}

func IsExportedOrBuiltinType(t reflect.Type) bool {
	return ast.IsExported(t.Name()) || t.PkgPath() == ""
}
