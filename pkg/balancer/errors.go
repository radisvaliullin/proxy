package balancer

import "fmt"

const (
	ErrKindClientNotConfig = iota
	ErrKindClientExceedLimti
	ErrKindCanNotGetUpstream
	ErrKindConfigWrongUpstr
)

var (
	ErrClientNotConfig   = BalancerError{Kind: ErrKindClientNotConfig}
	ErrClientExceedLimti = BalancerError{Kind: ErrKindClientExceedLimti}
	ErrCanNotGetUpstream = BalancerError{Kind: ErrKindCanNotGetUpstream}
	ErrConfigWrongUpstr  = BalancerError{Kind: ErrKindConfigWrongUpstr}
)

func getErrorMessage(kind int) string {
	switch kind {
	case ErrKindClientNotConfig:
		return "client not configured"
	case ErrKindClientExceedLimti:
		return "client has exceeded limit"
	case ErrKindCanNotGetUpstream:
		return "can not get next upstream"
	case ErrKindConfigWrongUpstr:
		return "config, wrong upstream address"
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
