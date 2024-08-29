package main

import (
	"context"
	"log"
	"net"
	"net/http"
	geeRPC2 "seven-day-web-framework/geeRPC2"
	"sync"
	"time"
)

func startServer(addrCh chan string) {
	var foo Foo
	l, err := net.Listen("tcp", ":8888")
	if err != nil {
		panic(err)
	}
	err = geeRPC2.Register(&foo)
	if err != nil {
		panic(err)
	}
	geeRPC2.HandleHTTP()
	addrCh <- l.Addr().String()
	_ = http.Serve(l, nil)
}

func call(addrCh chan string) {
	client, _ := geeRPC2.DialHTTP("tcp", <-addrCh)
	defer func() { _ = client.Close() }()

	time.Sleep(time.Second)
	// send request & receive response
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			args := &Args{Num1: i, Num2: i * i}
			var reply int
			if err := client.Call(context.Background(), "Foo.Sum", args, &reply); err != nil {
				log.Fatal("call Foo.Sum error:", err)
			}
			log.Printf("%d + %d = %d", args.Num1, args.Num2, reply)
		}(i)
	}
	wg.Wait()
}

func HTTPRPCStart() {
	log.SetFlags(0)
	ch := make(chan string)
	go call(ch)
	startServer(ch)
}
