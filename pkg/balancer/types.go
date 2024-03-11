package balancer

import "sync/atomic"

type clientBalance struct {
	// upstream addresses indexes (in balancer list of upstreams) limited by client perms
	// we need only indexes for fast look up
	upstrIdxs map[int]struct{}

	// conn limit (0 no limits)
	limit int
	// conn num counter, incremented atomically
	connCntr int32
}

func (c *clientBalance) incrClient() int {
	n := atomic.AddInt32(&c.connCntr, 1)
	return int(n)
}

func (c *clientBalance) decrClient() {
	_ = atomic.AddInt32(&c.connCntr, -1)
}

type upstrConnCntr struct {
	// counter
	cntr int
	// order position in order list
	orderIdx int
}

type Upstream interface {
	Addr() string
	Close()
}

type upstreamImpl struct {
	balancer  *Balancer
	clientId  string
	upstrAddr string
	upstrIdx  int
}

func (u *upstreamImpl) Addr() string {
	return u.upstrAddr
}

func (u *upstreamImpl) Close() {
	u.balancer.releaseUpstream(u.clientId, u.upstrIdx)
}
