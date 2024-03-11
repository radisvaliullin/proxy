package auth

type IAuth interface {
	AuthZ(string) bool
}
