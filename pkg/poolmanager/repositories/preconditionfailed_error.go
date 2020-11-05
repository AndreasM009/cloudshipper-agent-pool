package repositories

import "fmt"

// PreconditionFailedError error
type PreconditionFailedError struct {
	Err error
}

func (p *PreconditionFailedError) Error() string {
	return fmt.Sprintf("PreconditionFailed error: %s", p.Err)
}

// NewPreconditionFailedError new instance from error
func NewPreconditionFailedError(err error) error {
	return &PreconditionFailedError{
		Err: err,
	}
}
