package cacheCenter

import (
	"common/console"
	"encoding/binary"
	"fmt"
	"geeCache/byteview"
	"geeCache/consistentHash"
	"geeCache/keepalive"
	"geeCache/peer"
	"sync"
	"time"
)

const (
	keepalivePollTime   = 3000 * time.Second
	keepaliveCheckTimes = 3
)

type ICacheCenter interface {
	AddGetter(peer.NewGetter, ...string)
	DeleteGetter(...string)
	Get(string) (byteview.ByteView, error)
	keepalive.Keepalive
}

type CacheCenter struct {
	mu        sync.RWMutex
	peerMap   *consistentHash.Map
	getterMap map[string]peer.Getter
}

func New() *CacheCenter {
	cc := &CacheCenter{
		mu:        sync.RWMutex{},
		peerMap:   consistentHash.New(50, nil),
		getterMap: make(map[string]peer.Getter),
	}
	cc.run()
	return cc
}

func (cc *CacheCenter) AddGetter(newGetter peer.NewGetter, getterNames ...string) {
	cc.mu.Lock()
	for _, key := range getterNames {
		if _, ok := cc.getterMap[key]; !ok {
			cc.peerMap.Add(key)
			cc.getterMap[key] = newGetter(key)
		} else {
			console.Warn("[ cacheCenter ] %s is exist", key)
		}
	}
	cc.mu.Unlock()
}

func (cc *CacheCenter) DeleteGetter(getterNames ...string) {
	cc.mu.Lock()
	for _, key := range getterNames {
		if _, ok := cc.getterMap[key]; ok {
			cc.peerMap.Delete(key)
			delete(cc.getterMap, key)
		}
	}
	cc.mu.Unlock()
}

func (cc *CacheCenter) Get(key string) (byteview.ByteView, error) {
	cc.mu.RLock()
	peer := cc.peerMap.Get(key)
	if peer != "" {
		if getter, ok := cc.getterMap[peer]; ok {
			res, err := getter.Get(key)
			if err == nil {
				return byteview.New(res), nil
			}
			return byteview.ByteView{}, err
		}
	}

	var err error
	cc.mu.RUnlock()
	if peer == "" {
		err = fmt.Errorf("peer [%s] not exist", peer)
	}

	return byteview.ByteView{}, err
}

func (cc *CacheCenter) KeepAlive() {
	for key, getter := range cc.getterMap {
		go func(g peer.Getter) {
			for i := 0; i < keepaliveCheckTimes; i++ {
				res, err := g.Get(keepalive.KeepaliveStatusKey)
				if err == nil {
					status := binary.LittleEndian.Uint64(res)
					cc.disposeKeepaliveStatus(key, keepalive.KeepaliveStatus(status))
					return
				}
			}
			cc.info("%s is die", key)
			cc.DeleteGetter(key)
		}(getter)
	}
}

func (cc *CacheCenter) info(fmt string, args ...any) {
	console.Info("[CacheCenter] "+fmt, args...)
}

func (cc *CacheCenter) run() {
	go func() {
		for {
			time.Sleep(keepalivePollTime)
			cc.KeepAlive()
		}
	}()
}

func (cc *CacheCenter) disposeKeepaliveStatus(key string, status keepalive.KeepaliveStatus) {
	switch status {
	case keepalive.StatusOK:
		cc.info("%s is alive", key)
	default:
		cc.info("%s is unknown", key)
	}
}
