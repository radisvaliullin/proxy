package auth

var _ IAuth = (*Auth)(nil)

type Config struct {
	Clients []Client `yaml:"clients"`
}

// Client authentication works via TLS Certificates signed by root certificate
// Auth provides authorization (list available upstreams, limit of connections, etc)
type Auth struct {
	conf Config
}

func New(config Config) *Auth {
	a := &Auth{
		conf: config,
	}
	return a
}

// AuthN authenticate client
// most work done by mTLS, this method just check that client listed in auth config.
func (a *Auth) AuthN(clientId string) bool {
	for _, c := range a.conf.Clients {
		if c.Id == clientId {
			return true
		}
	}
	return false
}

// List all clients permissions
func (a *Auth) AllClientsPerms() Clients {
	return a.conf.Clients
}
