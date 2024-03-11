package proxy

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

	// in seconds
	// default value 10s
	HeartbeatTimeout int `yaml:"heartbeatTimeout"`
	// max number of bytes used for one read/write forward
	// value can be based on size of packet used in client/server protocol
	// default value 2048
	ForwardBuffSize int `yaml:"forwardBuffSize"`
}

func (c *Config) validate() error {
	if c.HeartbeatTimeout <= 0 {
		//
		c.HeartbeatTimeout = 10
	}
	if c.ForwardBuffSize <= 0 {
		c.ForwardBuffSize = 2048
	}
	return nil
}
