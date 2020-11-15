package voting

// UnauthorizedError represents a 404 error in a data, kv or object store
type UnauthorizedError struct {
	msg string
}

// Error returns the error message
func (err *UnauthorizedError) Error() string {
	return err.msg
}

// Is
func (err *UnauthorizedError) Is(other error) bool {
	_, ok := other.(*UnauthorizedError)
	return ok
}

// NewUnauthorizedError will return a UnauthorizedError object
func NewUnauthorizedError(msg string) *UnauthorizedError {
	return &UnauthorizedError{
		msg: msg,
	}
}

// VoterIDNotFoundError represents a 404 error in a data, kv or object store
type VoterIDNotFoundError struct {
	msg string
}

// Error returns the error message
func (err *VoterIDNotFoundError) Error() string {
	return err.msg
}

// Is
func (err *VoterIDNotFoundError) Is(other error) bool {
	_, ok := other.(*VoterIDNotFoundError)
	return ok
}

// NewVoterIDNotFoundError will return a VoterIDNotFoundError object
func NewVoterIDNotFoundError(msg string) *VoterIDNotFoundError {
	return &VoterIDNotFoundError{
		msg: msg,
	}
}

