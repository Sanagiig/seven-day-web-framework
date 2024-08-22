package geerpc

import (
	"context"
	"net"
	"testing"
	"time"
)

type NumPair struct {
	Num1 int
	Num2 int
}

type Calc struct{}

func (c *Calc) Add(argv NumPair, res *int) error {
	*res = argv.Num1 + argv.Num2
	return nil
}

func (c *Calc) Sub(argv NumPair, res *int) error {
	*res = argv.Num1 - argv.Num2
	return nil
}

func TestServerFind(t *testing.T) {
	c := new(Calc)
	err := Register(c)
	if err != nil {
		t.Fatalf("register err: %s", err.Error())
		return
	}

	svc, _, err := DefaultServer.findServer("Calc.Add")
	if err != nil {
		t.Fatalf("findServer err: %s", err.Error())
		return
	}

	if svc == nil {
		t.Fatalf("svc nil err")
		return
	}

	t.Logf("svc : %+v\n", svc)
}

func startServer(ch chan<- string, t *testing.T) *Server {
	lisn, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("server listen err \n %s", err.Error())
	}

	serv := NewServer()

	go serv.Accept(lisn)
	t.Log("server start.")
	ch <- lisn.Addr().String()
	return serv
}

func startClient(addr string, t *testing.T) *Client {
	t.Logf("client addr is [%s]", addr)
	client, err := Dial("tcp", addr, DefaultOption)
	if err != nil {
		t.Fatalf("client dial err \n %s", err.Error())
	}

	return client
}

func startServerAndClient(rsvc any, t *testing.T) (*Server, *Client) {
	ch := make(chan string, 1)
	server := startServer(ch, t)
	client := startClient(<-ch, t)

	err := server.Register(rsvc)
	if err != nil {
		t.Fatalf("server regist err \n %s", err.Error())
	}

	return server, client
}

func TestRpcCall(t *testing.T) {
	rcsv := new(Calc)
	_, client := startServerAndClient(rcsv, t)

	var reply int
	args := NumPair{Num1: 1, Num2: 2}
	ctx1, _ := context.WithTimeout(context.Background(), time.Second)
	err := client.Call(ctx1, "Calc.Add", args, &reply)
	if err != nil {
		t.Fatalf("call add err : %s", err.Error())
		return
	} else if reply != args.Num1+args.Num2 {
		t.Fatalf("Calc.Add 's reply err expect [%d] got [%d] ", args.Num1+args.Num2, reply)
		return
	}

	ctx2, _ := context.WithTimeout(context.Background(), time.Second)
	err = client.Call(ctx2, "Calc.Sub", args, &reply)
	if err != nil {
		t.Fatalf("call add err : %s", err.Error())
		return
	} else if reply != args.Num1-args.Num2 {
		t.Fatalf("Calc.Sub 's reply err expect [%d] got [%d] ", args.Num1-args.Num2, reply)
		return
	}
}
