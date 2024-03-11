package main

import (
	"log"

	"github.com/radisvaliullin/proxy/pkg/proxy"
)

func main() {

	p := proxy.New(proxy.Config{})
	if err := p.Start(); err != nil {
		log.Fatalf("main: proxy start: %v", err)
	}
}
