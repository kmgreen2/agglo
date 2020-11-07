package ticker

import (
	"github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/kvs"
)

type StreamStore interface {
	GetMessagesByName(name string) ImmutableMessage
	GetMessagesByTags(tags []string) ImmutableMessage
	GetMessageByUUID(uuid uuid.UUID) ImmutableMessage
	GetHistory(from uuid.UUID, to uuid.UUID) []ImmutableMessage
	GetHistoryToLastAnchor(uuid uuid.UUID) []ImmutableMessage
	Append(message ImmutableMessage) error
	// If the anchor is encoded into every message, then this is a simple fetch and compare from the Ticker: idx1 < idx2
	// In fact, that makes comparing cross-stream super easy
	HappenedBefore(lhs ImmutableMessage, rhs ImmutableMessage) (bool, error)
}

type KVStreamStore struct {
	kvStore kvs.KVStore
}