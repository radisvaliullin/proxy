package balancer

type IBalancer interface {
	Balance(string) (string, error)
	Close(string)
}
