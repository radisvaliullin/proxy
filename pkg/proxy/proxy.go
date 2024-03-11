package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net"
	"os"
)

type Config struct {
	// mTLS
	// Client CA Cert file path
	ClnCACertPath string `yaml:"clientCACertPath"`
	// Server Cert and Key file path
	SrvCertPath string `yaml:"serverCertPath"`
	SrvKeyPath  string `yaml:"serverKeyPath"`

	// Proxy Addr (ip/port)
	Addr string `yaml:"addr"`
	// Upstream Addr (ip/port)
	UpstreamAddr string `yaml:"upstreamAddr"`
}

type Proxy struct {
	config Config
}

func New(conf Config) *Proxy {
	p := &Proxy{
		config: conf,
	}
	return p
}

func (p *Proxy) Start() error {
	log.Print("proxy: start.")

	// Proxy mTLS certificates
	// client side certificate
	clnCaCertBytes, err := os.ReadFile(p.config.ClnCACertPath)
	if err != nil {
		log.Printf("proxy: start: read client CA cert file: %v", err)
		return err
	}
	clnCertPool := x509.NewCertPool()
	clnCertPool.AppendCertsFromPEM(clnCaCertBytes)
	// server side certificate
	srvCert, err := tls.LoadX509KeyPair(p.config.SrvCertPath, p.config.SrvKeyPath)
	if err != nil {
		log.Printf("proxy: start: server cert load: %v", err)
		return nil
	}

	// Proxy mTLS config
	srvMTLSConf := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		ClientCAs:                clnCertPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
		Certificates:             []tls.Certificate{srvCert},
	}

	ln, err := tls.Listen("tcp", p.config.Addr, srvMTLSConf)
	if err != nil {
		log.Printf("proxy: start: listen: %v", err)
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("proxy: start: accept tls conn: %v", err)
			continue
		}

		go p.handleConn(conn)
	}
}

func (p *Proxy) handleConn(conn net.Conn) {

	upstrmConn, err := net.Dial("tcp", p.config.UpstreamAddr)
	if err != nil {
		log.Printf("proxy: handler: upstream dial: %v", err)
		return
	}

	go func() {
		_, err := io.Copy(upstrmConn, conn)
		if err != nil {
			log.Printf("proxy: handler: forward conn to upstrmConn: %v", err)
			return
		}
	}()
	if _, err := io.Copy(conn, upstrmConn); err != nil {
		log.Printf("proxy: handler: forward upstrmConn to conn: %v", err)
		return
	}
}
