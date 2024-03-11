package balancer

import (
	"sync/atomic"

	"github.com/radisvaliullin/proxy/pkg/auth"
)

var _ IBalancer = (*Balancer)(nil)

type Config struct {
	UpstrmAddrs []string
}

type Balancer struct {
	conf Config

	// next upstream address index (if client not limited by client perms)
	// increment using atomic (see nextUpstrIdxIncr)
	// init -1 for start with 0 index
	nextUpstrIdx int64

	// next upstream address indexes by client (only for client limited by client perms)
	// map key client id
	// set only in New so no need protect with mutex
	clnUpstrIdx map[string]*clientNextUpstrIdx

	// Auth
	auth auth.IAuth
}

func New(config Config, iauth auth.IAuth) *Balancer {
	b := &Balancer{
		conf:         config,
		nextUpstrIdx: -1,
		clnUpstrIdx:  make(map[string]*clientNextUpstrIdx),
		auth:         iauth,
	}
	b.setClientUpstrIdx()
	return b
}

func (b *Balancer) setClientUpstrIdx() {
	// for each client with upstream perms build own next upstream idx list
	for _, client := range b.auth.AllClientsPerms() {
		clnNextUpstrIdx := &clientNextUpstrIdx{}
		for _, pu := range client.Perms.UpstreamAddrs {
			for i, u := range b.conf.UpstrmAddrs {
				if pu == u {
					clnNextUpstrIdx.upstrIdxs = append(clnNextUpstrIdx.upstrIdxs, i)
				}
			}
		}
		clnNextUpstrIdx.nextIdx = -1
		b.clnUpstrIdx[client.Id] = clnNextUpstrIdx
	}
}

// Balance return upstream address or error if client request denied
// Simple balancing
// if client do not have permission of upstreamAddrs each time return just next upstream address from pool
// if client has permission of upstreamAddrs return next address from pool limited by permission
func (b *Balancer) Balance(clientId string) (string, error) {
	return b.next(clientId), nil
}

func (b *Balancer) next(clientId string) string {
	var idx int
	// check if client has upstream perms
	if clnUpstrIdx, ok := b.clnUpstrIdx[clientId]; ok {
		idx = clnUpstrIdx.nextUpstrIdxIncr()
	} else {
		// case without client upstream perms
		idx = b.nextUpstrIdxIncr()
	}
	if idx < 0 {
		return ""
	}
	return b.conf.UpstrmAddrs[idx]
}

// Works only if idx range more less than (2^64)/2 because we should not overflow
// in our case number of upstreams can not be so big
// its tricky but it works
// it always return next value in range [0, max) atomically
// -1 means not index
func (b *Balancer) nextUpstrIdxIncr() int {
	max := len(b.conf.UpstrmAddrs)
	if max == 0 {
		return -1
	}
	max64 := int64(max)

	nextIdx64 := atomic.AddInt64(&b.nextUpstrIdx, 1)
	nextIdx := int(nextIdx64)
	if nextIdx < max {
		return nextIdx
	}
	nextIdx %= max
	if nextIdx == 0 {
		atomic.AddInt64(&b.nextUpstrIdx, -max64)
	}
	return nextIdx
}
