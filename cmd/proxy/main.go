package main

import (
	"log"

	"github.com/radisvaliullin/proxy/pkg/auth"
	"github.com/radisvaliullin/proxy/pkg/balancer"
	"github.com/radisvaliullin/proxy/pkg/config"
	"github.com/radisvaliullin/proxy/pkg/proxy"
)

func main() {

	// get config
	config, err := config.New()
	if err != nil {
		log.Fatalf("main: get config: %v", err)
	}
	log.Printf("main: config: %+v", config)

	// init dependencies
	au := auth.New(config.Auth)
	// balancer
	blnConf := balancer.Config{UpstrmAddrs: config.Proxy.UpstreamAddrs}
	blncer, err := balancer.New(blnConf, au)
	if err != nil {
		log.Fatalf("main: balancer: %v", err)
	}

	// init proxy and start
	p := proxy.New(config.Proxy, au, blncer)
	if err := p.Start(); err != nil {
		log.Fatalf("main: proxy start: %v", err)
	}
}
