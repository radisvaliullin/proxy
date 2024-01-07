package balancer

import "sync/atomic"

var _ IBalancer = (*Balancer)(nil)

type Config struct {
	UpstrmAddrs []string
}

type Balancer struct {
	conf Config

	// increment using atomic (see nextAddrIdxIncr)
	// init -1 for start with 0 index
	nextAddrIdx int64
}

func New(config Config) *Balancer {
	b := &Balancer{
		conf:        config,
		nextAddrIdx: -1,
	}
	return b
}

// Balance return upstream address or error if client request denied
// Simple balancing, each time return just next upstream address from pool
func (b *Balancer) Balance(clientId string) (string, error) {
	return b.next(), nil
}

func (b *Balancer) next() string {
	idx := b.nextAddrIdxIncr()
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
func (b *Balancer) nextAddrIdxIncr() int {
	max := len(b.conf.UpstrmAddrs)
	if max == 0 {
		return -1
	}
	max64 := int64(max)

	nextIdx64 := atomic.AddInt64(&b.nextAddrIdx, 1)
	nextIdx := int(nextIdx64)
	if nextIdx < max {
		return nextIdx
	}
	nextIdx %= max
	if nextIdx == 0 {
		atomic.AddInt64(&b.nextAddrIdx, -max64)
	}
	return nextIdx
}
