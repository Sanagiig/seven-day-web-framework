package main

import (
	"fmt"
	"net"
	geerpc "seven-day-web-framework/geeRPC2"
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

func startServer(ch chan<- string) *geerpc.Server {
	lisn, err := net.Listen("tcp", ":0")
	if err != nil {
		fmt.Printf("server listen err \n %s", err.Error())
	}

	serv := geerpc.NewServer()

	go serv.Accept(lisn)
	ch <- lisn.Addr().String()
	return serv
}

func startClient(addr string) *geerpc.Client {
	fmt.Printf("client addr is [%s]", addr)
	client, err := geerpc.Dial("tcp", addr, geerpc.DefaultOption)
	if err != nil {
		fmt.Printf("client dial err \n %s", err.Error())
	}

	return client
}

func startServerAndClient(rsvc any) (*geerpc.Server, *geerpc.Client) {
	ch := make(chan string, 1)
	server := startServer(ch)
	client := startClient(<-ch)

	err := server.Register(rsvc)
	if err != nil {
		fmt.Printf("server regist err \n %s", err.Error())
	}

	return server, client
}

func main() {
	rcsv := new(Calc)
	_, client := startServerAndClient(rcsv)

	var reply int
	args := NumPair{Num1: 1, Num2: 2}
	err := client.Call("Calc.Add", args, &reply)
	if err != nil {
		fmt.Printf("call add err : %s", err.Error())
		return
	} else if reply != args.Num1+args.Num2 {
		fmt.Printf("call.add 's reply err expect [%d] got [%d] ", args.Num1+args.Num2, reply)
		return
	}

}
