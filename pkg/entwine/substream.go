package entwine

import (
	"context"
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

func (ssa *SubStreamAppender) Append(ctx context.Context, message *UncommittedMessage) (gUuid.UUID, error) {
	anchorUuid, err := ssa.GetAnchorUuid()
	if err != nil {
		// ToDo(KMG): Assume the anchor is always set before calling append
		return gUuid.Nil, err
	}
	return ssa.streamStore.Append(ctx, message, ssa.subStreamID, anchorUuid)
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

func (ssa *SubStreamAppender) WriteLock() (context.Context, error) {
	return ssa.streamStore.SubStreamWriteLock(ssa.subStreamID)
}

func (ssa *SubStreamAppender) WriteUnlock(ctx context.Context) error {
	return ssa.streamStore.SubStreamWriteUnlock(ssa.subStreamID, ctx)
}

