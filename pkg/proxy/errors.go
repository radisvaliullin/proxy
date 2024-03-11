package proxy

import "fmt"

const (
	ErrKindForwardHeartBeat = iota
)

var (
	ErrForwardHeartBeat = ProxyError{Kind: ErrKindForwardHeartBeat}
)

func getErrorMessage(kind int) string {
	switch kind {
	case ErrKindForwardHeartBeat:
		return "forward heartbeat timeout"
	default:
		return "unknown"
	}
}

var _ error = ProxyError{}

type ProxyError struct {
	Kind int
}

func (e ProxyError) Error() string {
	return fmt.Sprintf("proxy error: %s", getErrorMessage(e.Kind))
}
