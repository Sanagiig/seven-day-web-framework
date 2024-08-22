package geeCache2

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/_geecache/"

type HttpPool struct {
	self     string
	basePath string
}

func NewHTTPPool(addr string) *HttpPool {
	return &HttpPool{
		self:     addr,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s \n", p.self, fmt.Sprintf(format, v...))
}

func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)

	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) < 2 {
		p.Log("[%s] bashPath not matched", r.URL.Path)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]
	group, ok := getGeeCacheFromGroup(groupName)
	if !ok {
		p.Log("[%s] not in cache group.", groupName)
		http.Error(w, fmt.Sprintf("[%s] not in cache group", groupName), http.StatusBadRequest)
		return
	}

	bv, err := group.Get(key)
	if err != nil {
		p.Log(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(bv.b)
}
