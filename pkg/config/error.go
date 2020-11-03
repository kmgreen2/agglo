package config

// Error is a generic error type for config errors
type Error struct {
	msg string
}

// NewConfigError is a constructor for config errors
func NewConfigError(msg string) *Error {
	return &Error{msg}
}

// Error returns a string describing the error
func (err *Error) Error() string {
	return err.msg
}
