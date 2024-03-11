package balancer

import "sync/atomic"

type clientBalance struct {
	// list of upstream addresses indexes (in balancer list of upstreams) limited by client perms
	upstrIdxs []int
	// idex of upstrIdxs list (init with -1 for start with 0)
	// incremented atomically
	nextUpstrIdxsIdx int32

	// conn limit (0 no limits)
	limit int
	// conn num counter, incremented atomically
	connCntr int32
}

// return index in general upstream addresses list (see balancer)
func (c *clientBalance) nextUpstrIdxIncr() int {
	max := len(c.upstrIdxs)
	idx := nextIdxIncr(max, &c.nextUpstrIdxsIdx)
	if idx < 0 {
		return -1
	}
	return c.upstrIdxs[idx]
}

func (c *clientBalance) incrClient() int {
	n := atomic.AddInt32(&c.connCntr, 1)
	return int(n)
}

func (c *clientBalance) decrClient() {
	_ = atomic.AddInt32(&c.connCntr, -1)
}
