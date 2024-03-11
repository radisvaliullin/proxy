package balancer

type IBalancer interface {
	Balance(string) (Upstream, error)
}
