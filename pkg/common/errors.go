package common

// ConflictError represents a conflict in a data, kv or object store
type ConflictError struct {
	msg string
}

// Error returns the error message
func (err *ConflictError) Error() string {
	return err.msg
}

// NewConflictError will return a ConflictError object
func NewConflictError(msg string) *ConflictError {
	return &ConflictError{
		msg: msg,
	}
}

// InternalError represents an internal error in a data, kv or object store
type InternalError struct {
	msg string
}

// Error returns the error message
func (err *InternalError) Error() string {
	return err.msg
}

// NewInternalError will return a InternalError object
func NewInternalError(msg string) *InternalError {
	return &InternalError{
		msg: msg,
	}
}

// NotFoundError represents a 404 error in a data, kv or object store
type NotFoundError struct {
	msg string
}

// Error returns the error message
func (err *NotFoundError) Error() string {
	return err.msg
}

// NewNotFoundError will return a NotFoundError object
func NewNotFoundError(msg string) *NotFoundError {
	return &NotFoundError{
		msg: msg,
	}
}

// InvalidError represents a 4xx error in a data, kv or object store
type InvalidError struct {
	msg string
}

// Error returns the error message
func (err *InvalidError) Error() string {
	return err.msg
}

// NewInvalidError will return a InvalidError object
func NewInvalidError(msg string) *InvalidError {
	return &InvalidError{
		msg: msg,
	}
}

// OutOfBoundsError represents a out-of-bounds error in a data, kv or object store
type OutOfBoundsError struct {
	msg string
}

// Error returns the error message
func (err *OutOfBoundsError) Error() string {
	return err.msg
}

// NewOutOfBoundsError will return a OutOfBoundsError object
func NewOutOfBoundsError(msg string) *OutOfBoundsError {
	return &OutOfBoundsError{
		msg: msg,
	}
}

// EndOfStreamError represents a end-of-stream error in a data, kv or object store
type EndOfStreamError struct {
	msg string
}

// Error returns the error message
func (err *EndOfStreamError) Error() string {
	return err.msg
}

// NewEndOfStreamError will return a EndOfStreamError object
func NewEndOfStreamError(msg string) *EndOfStreamError {
	return &EndOfStreamError{
		msg: msg,
	}
}
