package auth

type IAuth interface {
	AuthN(string) bool
	AllClientsPerms() Clients
}
