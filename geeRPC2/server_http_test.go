package geerpc

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func startHTTPServer(ch chan<- string, t *testing.T) *Server {
	const addr = ":8088"
	serv := NewServer()
	err := http.ListenAndServe(addr, serv)
	if err != nil {
		t.Fatalf("server listen err \n %s", err.Error())
	}

	t.Log("server start.")
	ch <- addr
	return serv
}

func startHTTPClient(addr string, t *testing.T) *Client {
	t.Logf("client addr is [%s]", addr)
	client, err := XDial(fmt.Sprintf("http@%s", addr))
	if err != nil {
		t.Fatalf("client dial err \n %s", err.Error())
	}

	return client
}

func startHttpServerAndClient(rsvc any, t *testing.T) (*Server, *Client) {
	ch := make(chan string, 1)
	sev := startHTTPServer(ch, t)
	client := startClient(<-ch, t)

	sev.Register(new(Calc))

	return sev, client
}

func TestRpcHTTPCall(t *testing.T) {
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
