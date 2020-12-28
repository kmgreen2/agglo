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

// Is
func (err *NotFoundError) Is(other error) bool {
	_, ok := other.(*NotFoundError)
	return ok
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

// SignatureError represents a end-of-stream error in a data, kv or object store
type SignatureError struct {
	msg string
}

// Error returns the error message
func (err *SignatureError) Error() string {
	return err.msg
}

// NewSignatureError will return a SignatureError object
func NewSignatureError(msg string) *SignatureError {
	return &SignatureError{
		msg: msg,
	}
}

// AlreadyCompletedError
type AlreadyCompletedError struct {
	msg string
}

// Error returns the error message
func (err *AlreadyCompletedError) Error() string {
	return err.msg
}

// NewAlreadyCompletedError will return a AlreadyCompletedError object
func NewAlreadyCompletedError(msg string) *AlreadyCompletedError {
	return &AlreadyCompletedError{
		msg: msg,
	}
}

// TimedOutError
type TimedOutError struct {
	msg string
}

// Error returns the error message
func (err *TimedOutError) Error() string {
	return err.msg
}

// NewTimedOutError will return a TimedOutError object
func NewTimedOutError(msg string) *TimedOutError {
	return &TimedOutError{
		msg: msg,
	}
}

