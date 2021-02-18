package entwine

import (
	"strings"
)

// SubStreamID is a wrapper for a SubStreamID
type SubStreamID string

// Equals will return true if the provided sub stream id is the same as this one
func (id SubStreamID) Equals(other SubStreamID) bool {
	return strings.Compare(string(id), string(other)) == 0
}
