package main

import (
	"log"

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

	// init proxy and start
	p := proxy.New(config.Proxy)
	if err := p.Start(); err != nil {
		log.Fatalf("main: proxy start: %v", err)
	}
}
