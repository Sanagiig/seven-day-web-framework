package main

import (
	"fmt"
	"log"
	"net/http"
	"seven-day-web-framework/geeCache2"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	geeCache2.NewGeeCache("scores", 2<<10, geeCache2.LocalGetterFn(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := geeCache2.NewHTTPPool(addr)
	log.Println("geecache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
