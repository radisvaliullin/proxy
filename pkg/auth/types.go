package auth

type Client struct {
	Id    string `yaml:"id"`
	Perms Perms  `yaml:"perms"`
}

type Perms struct {
	UpstreamAddrs []string `yaml:"upstreamAddrs"`
	Limit         int      `yaml:"limit"`
}

type Clients []Client
