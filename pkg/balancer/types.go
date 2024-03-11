package balancer

import "sync/atomic"

type clientNextUpstrIdx struct {
	// list of upstream addresses indexes limited by client perms
	upstrIdxs []int
	// idex of upstrIdxs list (init with -1 for start with 0)
	nextIdx int64
}

// see balancer nextUpstrIdxIncr for details
func (c *clientNextUpstrIdx) nextUpstrIdxIncr() int {
	max := len(c.upstrIdxs)
	if max == 0 {
		return -1
	}
	max64 := int64(max)

	nextIdx64 := atomic.AddInt64(&c.nextIdx, 1)
	nextIdx := int(nextIdx64)
	if nextIdx < max {
		return c.upstrIdxs[nextIdx]
	}
	nextIdx %= max
	if nextIdx == 0 {
		atomic.AddInt64(&c.nextIdx, -max64)
	}
	return c.upstrIdxs[nextIdx]
}
