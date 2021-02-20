package entwine

import (
	gUuid "github.com/google/uuid"
	"strings"
)

// SubStreamID is a wrapper for a SubStreamID
type SubStreamID string

// Equals will return true if the provided sub stream id is the same as this one
func (id SubStreamID) Equals(other SubStreamID) bool {
	return strings.Compare(string(id), string(other)) == 0
}

type SubStreamAppender struct {
	streamStore StreamStore
	subStreamID SubStreamID
}

func AllocateSubStreamID() SubStreamID {
	return SubStreamID(gUuid.New().String())
}

func NewSubStreamAppender(streamStore StreamStore, subStreamID SubStreamID) *SubStreamAppender {
	return &SubStreamAppender{
		streamStore: streamStore,
		subStreamID: subStreamID,
	}
}

func (ssa *SubStreamAppender) Head() (*StreamImmutableMessage, error) {
	return ssa.streamStore.Head(ssa.subStreamID)
}

func (ssa *SubStreamAppender) Append(message *UncommittedMessage, anchorTickerUuid gUuid.UUID) (gUuid.UUID, error) {
	return ssa.streamStore.Append(message, ssa.subStreamID, anchorTickerUuid)
}

func (ssa *SubStreamAppender) GetAnchorUuid() (gUuid.UUID, error) {
	return ssa.streamStore.GetCurrentAnchorUuid(ssa.subStreamID)
}

func (ssa *SubStreamAppender) SetAnchorUuid(uuid gUuid.UUID) error {
	return ssa.streamStore.SetCurrentAnchorUuid(ssa.subStreamID, uuid)
}

func (ssa *SubStreamAppender) GetHistory(startUuid, endUuid gUuid.UUID) ([]*StreamImmutableMessage, error) {
	return ssa.streamStore.GetHistory(startUuid, endUuid)
}

