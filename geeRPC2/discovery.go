/*
 * @Author: Sanagiig laiwenjunhhh@163.com
 * @Date: 2024-08-28 16:39:34
 * @LastEditors: Sanagiig laiwenjunhhh@163.com
 * @LastEditTime: 2024-08-28 17:21:33
 * @FilePath: \seven-day-web-framework\geeRPC2\discovery.go
 */
package geerpc

import (
	"errors"
	"math"
	"math/rand/v2"
	"sync"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	Refresh() error // refresh from remote registry
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServersDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServersDiscovery struct {
	r       *rand.Rand   // generate random number
	mu      sync.RWMutex // protect following
	servers []string
	index   int // record the selected position for robin algorithm
}

// NewMultiServerDiscovery creates a MultiServersDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServersDiscovery {
	md := &MultiServersDiscovery{
		servers: servers,
		r:       rand.New(rand.NewPCG(1, 5)),
	}
	md.index = md.r.IntN(math.MaxInt32 - 1)
	return md
}

// Get 根据模式获取一个 server string.
//
//	@receiver md
//	@param mode
//	@return string
//	@return error
func (md *MultiServersDiscovery) Get(mode SelectMode) (string, error) {
	md.mu.Lock()
	defer md.mu.Unlock()
	n := len(md.servers)
	if n == 0 {
		return "", errors.New("rpc discovery: no available servers")
	}
	switch mode {
	case RandomSelect:
		return md.servers[md.r.IntN(n)], nil
	case RoundRobinSelect:
		s := md.servers[md.index%n] // servers could be updated, so mode n to ensure safety
		md.index = (md.index + 1) % n
		return s, nil
	default:
		return "", errors.New("rpc discovery: not supported select mode")
	}
}

func (md *MultiServersDiscovery) GetAll() ([]string, error) {
	md.mu.RLock()
	defer md.mu.RUnlock()
	// return a copy of d.servers
	servers := make([]string, len(md.servers), len(md.servers))
	copy(servers, md.servers)
	return servers, nil
}

func (md *MultiServersDiscovery) Refresh() error {
	return nil
}

func (md *MultiServersDiscovery) Update(servers []string) error {
	md.mu.Lock()
	defer md.mu.Unlock()
	md.servers = servers
	return nil
}

var _ Discovery = (*MultiServersDiscovery)(nil)
