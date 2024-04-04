package cacheServer

import (
	"bytes"
	"common/console"
	"encoding/binary"
	"fmt"
	"geeCache"
	"geeCache/byteview"
	"geeCache/keepalive"
	"geeCache/peer"
	"log"
	"net/http"
	"strings"
	"sync"
)

const (
	DefaultCacheDir = "/_cache"
)

type CacheServer struct {
	mu         sync.RWMutex
	mainCache  *geeCache.Cache
	maxBytes   int64
	addr       string
	dir        string
	getterFunc peer.GetterFunc
}

func New(maxBytes int64, addr string, dir string, getterFunc peer.GetterFunc) *CacheServer {
	c := &CacheServer{
		mu:         sync.RWMutex{},
		mainCache:  geeCache.NewCache(maxBytes),
		maxBytes:   maxBytes,
		addr:       addr,
		dir:        dir,
		getterFunc: getterFunc,
	}

	if dir == "" {
		c.dir = DefaultCacheDir
	}
	return c
}

func (c *CacheServer) info(fmt string, args ...any) {
	console.Info("[CacheServer] "+fmt, args...)
}

func (c *CacheServer) Get(key string) []byte {
	value, ok := c.mainCache.Get(key)
	if ok {
		return value.ByteSlice()
	}
	return nil
}

func (c *CacheServer) Add(key string, value []byte) {
	c.mainCache.Add(key, byteview.New(value))
}

func (c *CacheServer) Run() {
	go func() {
		c.info("server started at [ %s ]", c.addr)
		log.Fatal(http.ListenAndServe(c.addr, c))
	}()
}

func (c *CacheServer) responseKeepalive(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	buff := bytes.NewBuffer([]byte{})
	err := binary.Write(buff, binary.LittleEndian, keepalive.StatusOK)
	if err != nil {
		return err
	}
	w.Write(buff.Bytes())
	return nil
}

func (c *CacheServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := r.URL.Path
	parts := strings.SplitN(path[1:], "/", 2)
	queryKey := parts[1]

	if strings.HasPrefix(path, DefaultCacheDir+"/"+keepalive.KeepaliveStatusKey) {
		err := c.responseKeepalive(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
		}
		return
	}

	if method != http.MethodGet || !strings.HasPrefix(path, c.dir) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("not found")))
		return
	}

	if queryKey == "" {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("key not found")))
		return
	}

	if c.getterFunc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("getter func is nil")))
		return
	}

	res, err := c.getterFunc(queryKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	} else {
		w.WriteHeader(http.StatusOK)
		i, e := w.Write(res)
		fmt.Println(i, e)
	}
}
