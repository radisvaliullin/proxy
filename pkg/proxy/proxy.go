package proxy

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"os"

	"github.com/radisvaliullin/proxy/pkg/auth"
	"github.com/radisvaliullin/proxy/pkg/balancer"
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
	// Upstream Addrs (list of ip:port)
	UpstreamAddrs []string `yaml:"upstreamAddrs"`
}

type Proxy struct {
	config Config

	auth   auth.IAuth
	blncer balancer.IBalancer
}

func New(conf Config, au auth.IAuth, blncer balancer.IBalancer) *Proxy {
	p := &Proxy{
		config: conf,
		auth:   au,
		blncer: blncer,
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
	log.Printf("proxy: handler: forward")
	defer log.Printf("proxy: handler: done")
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("proxy: handler: conn close: %v", err)
		}
	}()

	// auth connection
	clnId, err := p.authzConn(conn)
	if err != nil {
		log.Printf("proxy: handler: conn auth: %v", err)
		return
	}

	// get upstream address
	upstr, err := p.blncer.Balance(clnId)
	if err != nil {
		log.Printf("proxy: handler: conn balance, get upstream addr: %v", err)
		return
	}
	defer upstr.Close()

	// dial upstream
	upstrmConn, err := net.Dial("tcp", upstr.Addr())
	if err != nil {
		log.Printf("proxy: handler: upstream dial: %v", err)
		return
	}
	defer func() {
		if err := upstrmConn.Close(); err != nil {
			log.Printf("proxy: handler: upstream close: %v", err)
		}
	}()

	// cancel forward goroutines
	// if one of forward functions fail when context canceled
	// and conn handler should return and defer conn close
	fwdCtx, forwardCancel := context.WithCancel(context.Background())

	// forward conn->upstream and upstream->conn
	go func() {
		defer forwardCancel()
		if _, err := io.Copy(upstrmConn, conn); err != nil {
			log.Printf("proxy: handler: forward conn to upstrmConn: %v", err)
			return
		}
	}()
	go func() {
		defer forwardCancel()
		if _, err := io.Copy(conn, upstrmConn); err != nil {
			log.Printf("proxy: handler: forward upstrmConn to conn: %v", err)
			return
		}
	}()

	// lock until context canceled by one of forward goroutines
	<-fwdCtx.Done()
}

func (a *Proxy) authzConn(conn net.Conn) (string, error) {
	var (
		tc *tls.Conn
		ok bool
	)
	if tc, ok = conn.(*tls.Conn); !ok {
		return "", errors.New("tcp conn is not tls")
	}
	if err := tc.Handshake(); err != nil {
		log.Printf("proxy: handler: conn handshake: %v", err)
		return "", err
	}
	cs := tc.ConnectionState()
	if len(cs.PeerCertificates) <= 0 {
		log.Printf("proxy: handler: conn state, peer certificates not found")
		return "", errors.New("tls conn state, peer certificates not found")
	}
	id := cs.PeerCertificates[0].Subject.CommonName
	if !a.auth.AuthN(id) {
		return "", errors.New("tls conn, cert common name not authn")
	}

	return id, nil
}
