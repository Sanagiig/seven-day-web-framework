package service

import (
	"reflect"
	"sync/atomic"
)

type MethodType struct {
	method    reflect.Method
	ArgType   reflect.Type
	ReplyType reflect.Type
	numCalls  uint64
}

func (m *MethodType) NumCalls() uint64 {
	return atomic.LoadUint64(&m.numCalls)
}

func (m *MethodType) newArgv() reflect.Value {
	argv := getNewValue(m.ArgType)

	if m.ArgType.Kind() == reflect.Ptr {
		argv = reflect.New(m.ArgType.Elem())
	} else {
		argv = reflect.New(m.ArgType).Elem()
	}
	return argv
}

func (m *MethodType) NewArgv() reflect.Value {
	return m.newArgv()
}

func (m *MethodType) newReplyv() reflect.Value {
	replyv := getNewValue(m.ReplyType)

	switch m.ReplyType.Kind() {
	case reflect.Map:
		replyv.Elem().Set(reflect.MakeMap(m.ReplyType.Elem()))
	case reflect.Slice:
		replyv.Elem().Set(reflect.MakeSlice(m.ReplyType.Elem(), 0, 0))
	}
	return replyv
}

func (m *MethodType) NewReplyv() reflect.Value {
	return m.newReplyv()
}

func getNewValue(typ reflect.Type) reflect.Value {
	switch typ.Kind() {
	case reflect.Ptr:
		return reflect.New(typ.Elem())
	default:
		return reflect.New(typ).Elem()
	}
}
