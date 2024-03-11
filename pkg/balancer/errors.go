package balancer

import "fmt"

const (
	ErrKindClientNotConfig = iota
	ErrKindClientExceedLimti
	ErrKindCanNotGetUpstream
)

var (
	ErrClientNotConfig   = BalancerError{Kind: ErrKindClientNotConfig}
	ErrClientExceedLimti = BalancerError{Kind: ErrKindClientExceedLimti}
	ErrCanNotGetUpstream = BalancerError{Kind: ErrKindCanNotGetUpstream}
)

func getErrorMessage(kind int) string {
	switch kind {
	case ErrKindClientNotConfig:
		return "client not configured"
	case ErrKindClientExceedLimti:
		return "client has exceeded limit"
	case ErrKindCanNotGetUpstream:
		return "can not get next upstream"
	default:
		return "unknown"
	}
}

var _ error = BalancerError{}

type BalancerError struct {
	Kind int
}

func (e BalancerError) Error() string {
	return fmt.Sprintf("balancer error: %s", getErrorMessage(e.Kind))
}
