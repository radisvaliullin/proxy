package proxy

import "log"

type Proxy struct {
}

func New() *Proxy {
	p := &Proxy{}
	return p
}

func (p *Proxy) Start() {
	log.Print("proxy: start.")
}
