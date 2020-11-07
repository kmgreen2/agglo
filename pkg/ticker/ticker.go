package ticker

import (
	"github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/kvs"
)

type TickerStore interface {
	GetMessageByUUID(uuid uuid.UUID) ImmutableMessage
	GetHistory(from uuid.UUID, to uuid.UUID) []ImmutableMessage
	GetAnchor(proof []ImmutableMessage, subStreamID SubStreamID) ImmutableMessage
	HappenedBefore(lhs ImmutableMessage, rhs ImmutableMessage) (bool, error)
}

type KVTickerStore struct {
	kvStore kvs.KVStore
}
