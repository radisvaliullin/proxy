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
	nextUpstrIdx int32

	// clientsBalance stores balance parameters of clients
	// next upstream address indexes by client (only for client limited by client perms)
	// map key client id
	// set only in New so no need protect with mutex
	clientsBalance map[string]*clientBalance

	// Auth
	auth auth.IAuth
}

func New(config Config, iauth auth.IAuth) *Balancer {
	b := &Balancer{
		conf:           config,
		nextUpstrIdx:   -1,
		clientsBalance: make(map[string]*clientBalance),
		auth:           iauth,
	}
	b.setClientsBalanceParams()
	return b
}

func (b *Balancer) setClientsBalanceParams() {
	// for each client set client balance params struct
	// for client with upstream perms build own next upstream idx list
	for _, client := range b.auth.AllClientsPerms() {
		clnBlnc := &clientBalance{
			nextUpstrIdxsIdx: -1,
			limit:            client.Perms.Limit,
		}
		// set own upstream idx
		for _, pu := range client.Perms.UpstreamAddrs {
			for i, u := range b.conf.UpstrmAddrs {
				if pu == u {
					clnBlnc.upstrIdxs = append(clnBlnc.upstrIdxs, i)
				}
			}
		}
		b.clientsBalance[client.Id] = clnBlnc
	}
}

// Balance return upstream address or error if client request denied
// Simple balancing
// return next upstream address from general upstreams list
// or if client is limited by own permissions return next address from list of user upstreams
// check that client do not exceed limit
func (b *Balancer) Balance(clientId string) (string, error) {
	return b.balance(clientId)
}

// Close releases client from balancer stats
func (b *Balancer) Close(clientId string) {
	if clnBlnc, ok := b.clientsBalance[clientId]; ok {
		// if limit set
		if clnBlnc.limit > 0 {
			clnBlnc.decrClient()
		}
	}
}

func (b *Balancer) balance(clientId string) (upstr string, rerr error) {
	// use rerr only for defer
	// always return error value explicitly
	// return "", your_error

	var clnBlnc *clientBalance
	var ok bool
	// if not client balance (all registered client should have) then return empty
	if clnBlnc, ok = b.clientsBalance[clientId]; !ok {
		return "", ErrClientNotConfig
	}
	// decrement counter if incremented and error
	var isIncr bool
	defer func() {
		if rerr != nil && isIncr {
			clnBlnc.decrClient()

		}
	}()

	// check if client limit set
	if clnBlnc.limit > 0 {
		isIncr = true
		clnCnt := clnBlnc.incrClient()
		if clnCnt > clnBlnc.limit {
			return "", ErrClientExceedLimti
		}
	}

	// return next upstream address
	var idx int
	// check if client has upstream perms
	if len(clnBlnc.upstrIdxs) > 0 {
		idx = clnBlnc.nextUpstrIdxIncr()
	} else {
		// case without client upstream perms
		idx = b.nextUpstrIdxIncr()
	}
	if idx < 0 {
		return "", ErrCanNotGetUpstream
	}
	return b.conf.UpstrmAddrs[idx], nil
}

func (b *Balancer) nextUpstrIdxIncr() int {
	max := len(b.conf.UpstrmAddrs)
	return nextIdxIncr(max, &b.nextUpstrIdx)
}

// thread-safe incrementation using atomic
// works only if idx range more less than (2^32)/2 (positive range) because we should not overflow
// in our case number of upstreams can not be so big
// its tricky but it works
// it always return next value in range [0, max) atomically
// -1 means not index
func nextIdxIncr(max int, nextIdxPtr *int32) int {
	// devide 2 (half of positive range) to escape overflow
	if max <= 0 || max >= 2^31/2 {
		return -1
	}
	max32 := int32(max)

	nextIdx32 := atomic.AddInt32(nextIdxPtr, 1)
	nextIdx := int(nextIdx32)
	if nextIdx < max {
		return nextIdx
	}
	nextIdx %= max
	if nextIdx == 0 {
		atomic.AddInt32(nextIdxPtr, -max32)
	}
	return nextIdx
}
