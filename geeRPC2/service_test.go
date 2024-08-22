package geerpc

import (
	"fmt"
	"reflect"
	"testing"
)

type Foo int

type Args struct {
	Num1 int
	Num2 int
}

func (f Foo) Sum(arg Args, reply *int) error {
	*reply = arg.Num1 + arg.Num2
	return nil
}

// it's not a exported Method
func (f Foo) sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func _assert(condition bool, msg string, v ...interface{}) {
	if !condition {
		panic(fmt.Sprintf("assertion failed: "+msg, v...))
	}
}

func TestMethodExist(t *testing.T) {
	f := new(Foo)
	s := newService(f)
	methodNum := len(s.method)
	_assert(methodNum == 1, "method num err ,must be %d", 1)
	sum := s.method["Sum"]
	_assert(sum != nil, "Sum method not exist")
}

func TestMethodCall(t *testing.T) {
	f := new(Foo)
	s := newService(f)
	num1 := 1
	num2 := 2
	method := s.method["Sum"]
	argv := method.newArgv()
	replyv := method.newReplyv()

	argv.Set(reflect.ValueOf(Args{Num1: num1, Num2: num2}))
	err := s.call(method, argv, replyv)
	if err != nil {
		t.Fatalf("call err :%s", err.Error())
	}

	if int(replyv.Elem().Int()) != num1+num2 {
		t.Fatalf("call err :%s", err.Error())
	}
}
