package common

import (
	"github.com/pkg/errors"
	"fmt"
)

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// StringWithTrace will print an error string with stacktrace
func StringWithTrace(err error) string {
	msg := err.Error()
	if traceError, ok := err.(stackTracer); ok {
		msg += "\n"
		for _, f := range traceError.StackTrace() {
			msg += fmt.Sprintf("%+s:%d\n", f, f)
		}
	}
	return msg
}

// WithStack will decorate an error with a stack trace
func WithStack(err error) error {
	return errors.WithStack(err)
}

// Wrap will decorate an error with a stack trace and message
func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

// ConflictError represents a conflict in a Data, kv or object store
type ConflictError struct {
	msg string
}

// Error returns the error message
func (err *ConflictError) Error() string {
	return err.msg
}

// Is
func (err *ConflictError) Is(other error) bool {
	_, ok := other.(*ConflictError)
	return ok
}

// NewConflictError will return a ConflictError object
func NewConflictError(msg string) *ConflictError {
	return &ConflictError{
		msg: msg,
	}
}

// InternalError represents an internal error in a Data, kv or object store
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

// NotFoundError represents a 404 error in a Data, kv or object store
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

// InvalidError represents a 4xx error in a Data, kv or object store
type InvalidError struct {
	msg string
}

// Error returns the error message
func (err *InvalidError) Error() string {
	return err.msg
}

// Is
func (err *InvalidError) Is(other error) bool {
	_, ok := other.(*InvalidError)
	return ok
}

// NewInvalidError will return a InvalidError object
func NewInvalidError(msg string) *InvalidError {
	return &InvalidError{
		msg: msg,
	}
}

// OutOfBoundsError represents a out-of-bounds error in a Data, kv or object store
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

// EndOfStreamError represents a end-of-stream error in a Data, kv or object store
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

// SignatureError represents a end-of-stream error in a Data, kv or object store
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

// Is
func (err *TimedOutError) Is(other error) bool {
	_, ok := other.(*TimedOutError)
	return ok
}

// NewTimedOutError will return a TimedOutError object
func NewTimedOutError(msg string) *TimedOutError {
	return &TimedOutError{
		msg: msg,
	}
}

// CancelledError
type CancelledError struct {
}

// Error returns the error message
func (err *CancelledError) Error() string {
	return "Canceled"
}

// NewCancelledError will return a CancelledError object
func NewCancelledError() *CancelledError {
	return &CancelledError{
	}
}

// EvictedError represents an error condition where an object is unexpectedly not found
type EvictedError struct {
	msg string
}

// NewEvictedError is the constructor for EvictedError
func NewEvictedError(msg string) *EvictedError {
	return &EvictedError{msg}
}

// Error returns the string representation of the EvictedError
func (e *EvictedError) Error() string {
	return e.msg
}

// InconsistentStateError represents an error condition where an object is unexpectedly not found
type InconsistentStateError struct {
	msg string
}

// NewInconsistentStateError is the constructor for InconsistentStateError
func NewInconsistentStateError(msg string) *InconsistentStateError {
	return &InconsistentStateError{msg}
}

// Error returns the string representation of the InconsistentStateError
func (e *InconsistentStateError) Error() string {
	return e.msg
}

// AlreadyOpenedError represents an error condition when a persistent resource is already opened by another process
type AlreadyOpenedError struct {
	msg string
}

// NewAlreadyOpenedError is the constructor for AlreadyOpenedError
func NewAlreadyOpenedError(msg string) *AlreadyOpenedError {
	return &AlreadyOpenedError{msg}
}

// Error returns the string representation of the AlreadyOpenedError
func (e *AlreadyOpenedError) Error() string {
	return e.msg
}

// ClosedError represents an error condition when a persistent resource is closed
type ClosedError struct {
	msg string
}

// NewClosedError is the constructor for ClosedError
func NewClosedError(msg string) *ClosedError {
	return &ClosedError{msg}
}

// Error returns the string representation of the ClosedError
func (e *ClosedError) Error() string {
	return e.msg
}

// EmptyQueue represents a condition when a queue is empty
type EmptyQueue struct {
	msg string
}

// NewEmptyQueue is the constructor for EmptyQueue
func NewEmptyQueue(msg string) *EmptyQueue {
	return &EmptyQueue{msg}
}

// Error returns the string representation of the EmptyQueue
func (e *EmptyQueue) Error() string {
	return e.msg
}

// Is
func (err *EmptyQueue) Is(other error) bool {
	_, ok := other.(*EmptyQueue)
	return ok
}

