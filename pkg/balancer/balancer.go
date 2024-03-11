package balancer

import (
	"sync"

	"github.com/radisvaliullin/proxy/pkg/auth"
)

var _ IBalancer = (*Balancer)(nil)

type Config struct {
	UpstrmAddrs []string
}

func (c *Config) validate() error {
	if len(c.UpstrmAddrs) <= 0 {
		return ErrConfigWrongUpstr
	}
	return nil
}

// Balancer balance using a least connection method
// (previous version used round-robin, see history of commits)
type Balancer struct {
	conf Config

	// general mutex for object used to get next upstream address
	upstrMx sync.Mutex
	// stores number of connections of upstream address
	upstrConnCntr []upstrConnCntr
	// upstream indexes (in list of upstreams) ordered from less conn to max conn number
	upstrIdxsByConnNum []int

	// clientsBalance stores balance parameters of clients
	// upstream address indexes by client (only for client limited by client perms)
	// map key client id
	// set only in New so no need protect with mutex
	clientsBalance map[string]*clientBalance

	// Auth
	auth auth.IAuth
}

func New(config Config, iauth auth.IAuth) (*Balancer, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}
	b := &Balancer{
		conf:               config,
		upstrConnCntr:      make([]upstrConnCntr, len(config.UpstrmAddrs)),
		upstrIdxsByConnNum: make([]int, len(config.UpstrmAddrs)),
		clientsBalance:     make(map[string]*clientBalance),
		auth:               iauth,
	}
	b.setBalancerParams()
	return b, nil
}

func (b *Balancer) setBalancerParams() {
	// set default values of upstreams idxs order
	// in initial state all conn number is zero so we just order in same way as in in config
	for i := 0; i < len(b.conf.UpstrmAddrs); i++ {
		b.upstrConnCntr[i].orderIdx = i
		b.upstrIdxsByConnNum[i] = i
	}
	// for each client set client balance params struct
	// for client with upstream perms build own upstream idx list
	for _, client := range b.auth.AllClientsPerms() {
		clnBlnc := &clientBalance{
			limit:     client.Perms.Limit,
			upstrIdxs: make(map[int]struct{}, len(client.Perms.UpstreamAddrs)),
		}
		// set own upstream idx
		for _, pu := range client.Perms.UpstreamAddrs {
			for i, u := range b.conf.UpstrmAddrs {
				if pu == u {
					// we need only indexes
					clnBlnc.upstrIdxs[i] = struct{}{}
				}
			}
		}
		b.clientsBalance[client.Id] = clnBlnc
	}
}

// Balance return upstream interface or error if client request denied
// Simple balancing
// return next upstream address usign a least connection method
// if client is limited by own permissions return next address from list of client upstreams
// check that client do not exceed limit
func (b *Balancer) Balance(clientId string) (Upstream, error) {
	return b.balance(clientId)
}

// releases client from balancer stats
func (b *Balancer) releaseUpstream(clientId string, upstrIdx int) {
	if clnBlnc, ok := b.clientsBalance[clientId]; ok {
		// if limit set
		if clnBlnc.limit > 0 {
			clnBlnc.decrClient()
		}
	}
	b.decrUpstr(upstrIdx)
}

func (b *Balancer) balance(clientId string) (upstr Upstream, rerr error) {
	// use rerr only for defer
	// always return error value explicitly
	// return nil, error

	var clnBlnc *clientBalance
	// if not client balance (all registered client should have) then return empty
	if cb, ok := b.clientsBalance[clientId]; ok {
		clnBlnc = cb
	} else {
		return nil, ErrClientNotConfig
	}
	// decrement counter if incremented and error
	var isClnCntrIncr bool
	defer func() {
		if rerr != nil && isClnCntrIncr {
			clnBlnc.decrClient()
		}
	}()

	// check if client limit set
	if clnBlnc.limit > 0 {
		isClnCntrIncr = true
		clnCnt := clnBlnc.incrClient()
		if clnCnt > clnBlnc.limit {
			return nil, ErrClientExceedLimti
		}
	}

	// next upstream address
	upstrIdx := b.nextUpstreamIdx(clnBlnc)

	// upstream
	upstr = &upstreamImpl{
		balancer:  b,
		clientId:  clientId,
		upstrAddr: b.conf.UpstrmAddrs[upstrIdx],
		upstrIdx:  upstrIdx,
	}
	return upstr, nil
}

func (b *Balancer) nextUpstreamIdx(clnBalance *clientBalance) int {
	b.upstrMx.Lock()
	defer b.upstrMx.Unlock()

	upstrIdx := b.upstrIdxsByConnNum[0]
	// if we have client specific permition list limit idx by the list
	if len(clnBalance.upstrIdxs) > 0 {
		for i := 0; i < len(b.upstrIdxsByConnNum); i++ {
			upstrIdx = b.upstrIdxsByConnNum[i]
			// at least one should be in map
			if _, ok := clnBalance.upstrIdxs[upstrIdx]; ok {
				break
			}
		}
	}

	// update
	b.incrUpstrCntrNotSafe(upstrIdx)

	return upstrIdx
}

func (b *Balancer) decrUpstr(upstrIdx int) {
	b.upstrMx.Lock()
	defer b.upstrMx.Unlock()
	b.decrUpstrCntrNotSafe(upstrIdx)
}

// not thread-safe
func (b *Balancer) incrUpstrCntrNotSafe(upstrIdx int) {
	// update upstream counter
	b.upstrConnCntr[upstrIdx].cntr++

	// update order
	// if current element in order list higher than next need swap
	// repead for next elements
	orderIdx := b.upstrConnCntr[upstrIdx].orderIdx
	for i := orderIdx; i < (len(b.upstrIdxsByConnNum) - 1); i++ {
		upstrIdx := b.upstrIdxsByConnNum[i]
		nextUpstrIdx := b.upstrIdxsByConnNum[i+1]
		if b.upstrConnCntr[upstrIdx].cntr > b.upstrConnCntr[nextUpstrIdx].cntr {
			// swap
			b.upstrIdxsByConnNum[i], b.upstrIdxsByConnNum[i+1] = b.upstrIdxsByConnNum[i+1], b.upstrIdxsByConnNum[i]
			b.upstrConnCntr[upstrIdx].orderIdx = i + 1
			b.upstrConnCntr[nextUpstrIdx].orderIdx = i
		} else {
			break
		}
	}
}

// not thread-safe
func (b *Balancer) decrUpstrCntrNotSafe(upstrIdx int) {
	// update upstream counter
	b.upstrConnCntr[upstrIdx].cntr--

	// update order
	// if current element in order list lower than prev need swap
	// repead for prev elements
	orderIdx := b.upstrConnCntr[upstrIdx].orderIdx
	for i := orderIdx; i > 0; i-- {
		upstrIdx := b.upstrIdxsByConnNum[i]
		prevUpstrIdx := b.upstrIdxsByConnNum[i-1]
		if b.upstrConnCntr[upstrIdx].cntr < b.upstrConnCntr[prevUpstrIdx].cntr {
			// swap
			b.upstrIdxsByConnNum[i], b.upstrIdxsByConnNum[i-1] = b.upstrIdxsByConnNum[i-1], b.upstrIdxsByConnNum[i]
			b.upstrConnCntr[upstrIdx].orderIdx = i - 1
			b.upstrConnCntr[prevUpstrIdx].orderIdx = i
		} else {
			break
		}
	}
}
