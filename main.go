package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"geeCache/cacheCenter"
	"geeCache/cacheServer"
	"geeCache/peer"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"net/http"
	"seven-day-web-framework/geeRPC"
	"seven-day-web-framework/geeRPC/client"
	"seven-day-web-framework/geeRPC/codec"
	"seven-day-web-framework/protobuf/user"
	"sync"
	"time"
)

func httpPoolTest() {

}

func peerTest() {
	db := map[string]string{
		"tom":  "630",
		"Jack": "589",
		"Sam":  "567",
	}

	apiAddr := "localhost:9999"
	cacheAddr := []string{"http://localhost:8001", "http://localhost:8002", "http://localhost:8003"}

	createCacheCenter := func() *cacheCenter.CacheCenter {
		cc := cacheCenter.New()
		cc.AddGetter(peer.NewHttpGetter, cacheAddr...)
		return cc
	}

	createCache := func() {
		for _, addr := range cacheAddr {
			cache := cacheServer.New(10000, addr[7:], "", func(key string) ([]byte, error) {
				val, ok := db[key]
				if !ok {
					return nil, fmt.Errorf("db [%s] not found", key)
				}
				return []byte(val), nil
			})
			cache.Run()
		}
	}

	createApiServer := func(cc *cacheCenter.CacheCenter) {
		http.Handle("/api", http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				key := r.URL.Query().Get("key")
				get, err := cc.Get(key)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(err.Error()))
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write(get.ByteSlice())
			},
		))
		http.ListenAndServe(apiAddr, nil)
	}

	cc := createCacheCenter()
	createCache()
	createApiServer(cc)
}

func testProto() {
	test := &user.Student{
		Name:   "geektutu",
		Male:   true,
		Scores: []int32{98, 85, 88},
	}
	data, err := proto.Marshal(test)
	if err != nil {
		log.Fatal("marshaling error: ", err)
	}
	newTest := &user.Student{}
	err = proto.Unmarshal(data, newTest)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	// Now test and newTest contain the same data.
	if test.GetName() != newTest.GetName() {
		log.Fatalf("data mismatch %q != %q", test.GetName(), newTest.GetName())
	}
}

func testSqlite() {
	db, _ := sql.Open("sqlite3", "gee.db")
	defer func() { _ = db.Close() }()
	_, _ = db.Exec("DROP TABLE IF EXISTS User;")
	_, _ = db.Exec("CREATE TABLE User(Name text);")
	result, err := db.Exec("INSERT INTO User(`Name`) values (?), (?)", "Tom", "Sam")
	if err == nil {
		affected, _ := result.RowsAffected()
		log.Println(affected)
	}
	row := db.QueryRow("SELECT Name FROM User LIMIT 1")
	var name string
	if err := row.Scan(&name); err == nil {
		log.Println(name)
	}
}

func rpcConnTest() {
	startServer := func(addr chan string) {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal("network error:", err)
		}
		log.Println("start rpc server on", l.Addr())
		addr <- l.Addr().String()
		geeRPC.Accept(l)
	}

	addr := make(chan string)
	go startServer(addr)

	conn, _ := net.Dial("tcp", <-addr)
	defer func() { _ = conn.Close() }()

	time.Sleep(time.Second)
	_ = json.NewEncoder(conn).Encode(geeRPC.DefaultOption)
	cc := codec.NewGobCodec(conn)

	for i := 0; i < 5; i++ {
		h := &codec.Header{
			ServiceMethod: "Foo.Sum",
			Seq:           uint64(i),
		}
		_ = cc.Write(h, fmt.Sprintf("geerpc req %d", h.Seq))
		_ = cc.ReadHeader(h)
		var reply string
		_ = cc.ReadBody(&reply)
		log.Println("reply", reply)
	}
}

type Foo int

type Args struct{ Num1, Num2 int }

func (f Foo) Sum(args Args, reply *int) error {
	*reply = args.Num1 + args.Num2
	return nil
}

func testRpcService() {
	startServer := func(addr chan string) {
		var foo Foo
		if err := geeRPC.Register(&foo); err != nil {
			log.Fatal("register error:", err)
		}
		// pick a free port
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			log.Fatal("network error:", err)
		}
		log.Println("start rpc server on", l.Addr())
		addr <- l.Addr().String()
		geeRPC.Accept(l)
		return
	}

	log.SetFlags(0)
	addr := make(chan string)
	go startServer(addr)
	client, _ := client.Dial("tcp", <-addr)
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

func main() {
	testRpcService()
}
