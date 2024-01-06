package auth

var _ IAuth = (*Auth)(nil)

type Config struct {
	Clients []Client `yaml:"clients"`
}

type Client struct {
	Id string `yaml:"id"`
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

// AuthZ permissions
// right now just authorize any user for all upstreams
func (a *Auth) AuthZ(clientId string) bool {
	for _, c := range a.conf.Clients {
		if c.Id == clientId {
			return true
		}
	}
	return false
}
