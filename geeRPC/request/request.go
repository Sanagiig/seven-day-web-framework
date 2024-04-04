package request

import (
	"reflect"
	"seven-day-web-framework/geeRPC/codec"
	"seven-day-web-framework/geeRPC/service"
)

// request stores all information of a call
type Request struct {
	H            *codec.Header // header of request
	Argv, Replyv reflect.Value // argv and replyv of request
	Service      *service.Service
	MethodType   *service.MethodType
}
