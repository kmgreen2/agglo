package entwine

import (
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"strings"
	"sync"
	"time"
)

/*
	Internal key format :

	object metadata: serialized StreamImmutableMessage
	Primary record: <UUID>-n -> <data or data descriptor>, <object metadata>, <tags>
	Previous node:  <UUID>-p -> <previous UUID>
	Anchor node:  <UUID>-a -> <anchor UUID>
 	Name (listing): md5(name)[:4]-<name>-<UUID>
 	Tags: md5(tag)[:4]-<tag>-<UUID>
 */

// StreamStore is the interface for a converged stream of substreams
type StreamStore interface {
	GetMessagesByName(name string) ([]*StreamImmutableMessage, error)
	GetMessagesByTags(tags []string) ([]*StreamImmutableMessage, error)
	GetMessageByUUID(uuid gUuid.UUID) (*StreamImmutableMessage, error)
	GetHistory(start gUuid.UUID, end gUuid.UUID) ([]*StreamImmutableMessage, error)
	GetHistoryToLastAnchor(uuid gUuid.UUID) ([]*StreamImmutableMessage, error)
	Append(message *UncommittedMessage, subStreamID SubStreamID, anchorTickerUuid gUuid.UUID) (gUuid.UUID, error)
	Head(subStreamID SubStreamID) (*StreamImmutableMessage, error)
	Create(subStreamID SubStreamID, digestType common.DigestType,
		signer crypto.Signer, anchorTickerUuid gUuid.UUID) error

	// This function needs to be at a higher level
	// If the anchor is encoded into every message, then this is a simple fetch and compare from the Ticker: idx1 < idx2
	// In fact, that makes comparing cross-stream super easy
	//HappenedBefore(lhs *StreamImmutableMessage, rhs *StreamImmutableMessage) (bool, error)
}

// KVStreamStore is an implementation of StreamStore that is backed by an in-memory map
type KVStreamStore struct {
	kvStore kvs.KVStore
	heads map[string]*StreamImmutableMessage
	digestType common.DigestType
	writeLocks map[string]*sync.Mutex
	streamLock *sync.Mutex
}

// NewKVStreamStore returns a new KVStreamStore backed by the provided KVStore
// ToDo(KMG): Need to init heads, write locks from state in the backing KVStore
func NewKVStreamStore(kvStore kvs.KVStore, digestType common.DigestType) *KVStreamStore {
	return &KVStreamStore{
		kvStore: kvStore,
		digestType: digestType,
		heads: make(map[string]*StreamImmutableMessage),
		writeLocks: make(map[string]*sync.Mutex),
		streamLock: &sync.Mutex{},
	}
}

// Head will return the latest message appended to the ticker stream
func (streamStore *KVStreamStore) Head(subStreamID SubStreamID) (*StreamImmutableMessage, error) {
	if head, ok := streamStore.heads[string(subStreamID)]; ok {
		return head, nil
	}
	return nil, common.NewNotFoundError(fmt.Sprintf("Head - substream not found: %s", subStreamID))
}

// Create will create a new substream
func (streamStore *KVStreamStore) Create(subStreamID SubStreamID, digestType common.DigestType,
	signer crypto.Signer, anchorTickerUuid gUuid.UUID) error {
	genesisMessage, err := NewStreamGenesisMessage(subStreamID, digestType, signer,
		anchorTickerUuid)
	if err != nil {
		return err
	}
	err = streamStore.append(genesisMessage)
	if err != nil {
		return err
	}
	streamStore.heads[string(subStreamID)] = genesisMessage
	return nil
}

// GetMessagesByName will return all messages that have a given name; otherwise, return an error
func (streamStore *KVStreamStore) GetMessagesByName(name string) ([]*StreamImmutableMessage, error) {
	namePrefix, err := NameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	keys, err := streamStore.kvStore.List(namePrefix)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, common.NewNotFoundError(fmt.Sprintf("GetMessagesByName - not found: %s", name))
	}

	messages := make([]*StreamImmutableMessage, len(keys))
	for i, _ := range keys {
		messageUuid, err := UuidFromNameKey(keys[i])
		if err != nil {
			return nil, err
		}
		messageBytes, err := streamStore.kvStore.Get(messageUuid)
		if err != nil {
			return nil, err
		}
		messages[i], err = NewStreamImmutableMessageFromBuffer(messageBytes)
		if err != nil {
			return nil, err
		}
	}
	return messages, nil
}

// GetMessagesByTags will return all messages that have a given set of tags; otherwise, return an error
func (streamStore *KVStreamStore) GetMessagesByTags(tags []string) ([]*StreamImmutableMessage, error) {
	var allKeys []string

	for _, tag := range tags {
		tagPrefix, err := TagKeyPrefix(tag)
		if err != nil {
			return nil, err
		}
		keys, err := streamStore.kvStore.List(tagPrefix)
		if err != nil {
			return nil, err
		}
		allKeys = append(allKeys, keys...)
	}

	if len(allKeys) == 0 {
		return nil, common.NewNotFoundError(fmt.Sprintf("GetMessagesByTags - not found: %v", tags))
	}

	messages := make([]*StreamImmutableMessage, len(allKeys))
	for i, _ := range allKeys {
		messageUuid, err := UuidFromTagKey(allKeys[i])
		if err != nil {
			return nil, err
		}
		messageBytes, err := streamStore.kvStore.Get(messageUuid)
		if err != nil {
			return nil, err
		}
		messages[i], err = NewStreamImmutableMessageFromBuffer(messageBytes)
		if err != nil {
			return nil, err
		}
	}

	return messages, nil
}

// GetMessageByUUID will return the message with a given UUID; otherwise, return an error
func (streamStore *KVStreamStore) GetMessageByUUID(uuid gUuid.UUID) (*StreamImmutableMessage, error) {
	messageBytes, err := streamStore.kvStore.Get(uuid.String())
	if err != nil {
		return nil, err
	}
	return NewStreamImmutableMessageFromBuffer(messageBytes)
}

// GetMessages will return the messages for the given UUIDs; otherwise, return an error
func (streamStore *KVStreamStore) GetMessages(uuids []gUuid.UUID) ([]*StreamImmutableMessage, error) {
	var chainedMessages []*StreamImmutableMessage
	for _, myUuid := range uuids {
		messageBytes, err := streamStore.kvStore.Get(myUuid.String())
		if err != nil {
			return nil, err
		}
		message, err := NewStreamImmutableMessageFromBuffer(messageBytes)
		if err != nil {
			return nil, err
		}
		chainedMessages = append(chainedMessages, message)
	}
	return chainedMessages, nil
}

// GetHistory will return the ordered, immutable history between two UUIDs; otherwise return an error
func (streamStore *KVStreamStore) GetHistory(start gUuid.UUID, end gUuid.UUID) ([]*StreamImmutableMessage, error) {
	var chainedUuids []gUuid.UUID
	curr := end

	for {
		// If a nil start UUID is given, then process the entire history back to the genesis block
		if curr == gUuid.Nil && start == gUuid.Nil {
			break
		}
		if err := streamStore.kvStore.Head(curr.String()); err != nil {
			return nil, err
		}

		chainedUuids = append(chainedUuids, curr)
		if strings.Compare(curr.String(), start.String()) == 0 {
			break
		}
		prevBytes, err := streamStore.kvStore.Get(PreviousNodeKey(curr))

		// No previous message, assumes we have reached the first
		// ToDo(KMG): Do we care?  Should we check the first message and return an error?
		if errors.Is(err, &common.NotFoundError{}) {
			break
		} else if err != nil {
			return nil, err
		}
		prev, err := BytesToUUID(prevBytes)
		if err != nil {
			return nil, err
		}
		curr = prev
	}
	messages, err := streamStore.GetMessages(chainedUuids)
	if err != nil {
		return nil, err
	}

	ReverseStreamMessages(messages)

	return messages, nil
}

// getAnchorUUID will return the anchor UUID for a give message; otherwise, return an error
func (streamStore *KVStreamStore) getAnchorUUID(uuid gUuid.UUID) (gUuid.UUID, error) {
	anchorBytes, err := streamStore.kvStore.Get(AnchorNodeKey(uuid))
	if err != nil {
		return gUuid.Nil, err
	}
	anchorUuid, err := BytesToUUID(anchorBytes)
	if err != nil {
		return gUuid.Nil, err
	}
	return anchorUuid, nil
}

// GetHistoryToLastAnchor will return the immutable history from the provided UUID back to the last message with
// a different anchor; otherwise, return an error
func (streamStore *KVStreamStore) GetHistoryToLastAnchor(uuid gUuid.UUID) ([]*StreamImmutableMessage, error) {
	var chainedUuids []gUuid.UUID

	curr := uuid
	currAnchor, err := streamStore.getAnchorUUID(curr)
	if err != nil {
		return nil, err
	}
	chainedUuids = append(chainedUuids, curr)
	for {
		if err := streamStore.kvStore.Head(curr.String()); err != nil {
			return nil, err
		}

		prevBytes, err := streamStore.kvStore.Get(PreviousNodeKey(curr))
		// No previous message, assumes we have reached the first
		// ToDo(KMG): Do we care?  Should we check the first message and return an error?
		if errors.Is(err, &common.NotFoundError{}) {
			break
		} else if err != nil {
			return nil, err
		}
		prev, err := BytesToUUID(prevBytes)
		if err != nil {
			return nil, err
		}
		prevAnchor, err := streamStore.getAnchorUUID(prev)
		if err != nil {
			return nil, err
		}
		chainedUuids = append(chainedUuids, prev)
		if strings.Compare(currAnchor.String(), prevAnchor.String()) != 0 {
			break
		}
		curr = prev
		currAnchor = prevAnchor
	}
	return streamStore.GetMessages(chainedUuids)
}

// getHead will return the head of a specific substream; otherwise, return an error
func (streamStore *KVStreamStore) getHead(subStreamID SubStreamID) (*StreamImmutableMessage, error) {
	if head, ok := streamStore.heads[string(subStreamID)]; ok {
		return head , nil
	}
	return nil, common.NewNotFoundError(fmt.Sprintf("GetHead - cannot find substream: %s", subStreamID))
}

// setHead will set the head of a specific substream; otherwise, return an error
func (streamStore *KVStreamStore) setHead(subStreamID SubStreamID, message *StreamImmutableMessage) error {
	streamStore.heads[string(subStreamID)] = message
	return nil
}

// Append will append an uncommitted message to a substream, anchored at the provided anchor UUID; otherwise return
// an error.
func (streamStore *KVStreamStore) Append(message *UncommittedMessage, subStreamID SubStreamID,
	anchorTickerUuid gUuid.UUID) (gUuid.UUID, error) {
	ts := time.Now().Unix()

	streamStore.streamLock.Lock()
	if _, ok := streamStore.writeLocks[string(subStreamID)]; !ok {
		streamStore.writeLocks[string(subStreamID)] = &sync.Mutex{}
	}
	streamStore.streamLock.Unlock()

	// Take lock
	streamStore.writeLocks[string(subStreamID)].Lock()
	defer streamStore.writeLocks[string(subStreamID)].Unlock()

	head, err := streamStore.getHead(subStreamID)
	if err != nil {
		return gUuid.Nil, err
	}
	immutableMessage, err := NewStreamImmutableMessage(subStreamID, message.objectDescriptor, message.name,
		message.tags, streamStore.digestType, message.signer, ts, head, anchorTickerUuid)
	if err != nil {
		return gUuid.Nil, err
	}

	err = streamStore.append(immutableMessage)
	if err != nil {
		return gUuid.Nil, err
	}

	return immutableMessage.uuid, nil
}

// append preps and performs all of the storage operations for appending a new message
// ToDo(KMG): Need to rollback any incomplete append
func (streamStore *KVStreamStore) append(message *StreamImmutableMessage) error {
	newUuid := message.uuid

	prevUuidBytes, err := UuidToBytes(message.Prev())
	if err != nil {
		return err
	}

	anchorUuidBytes, err := UuidToBytes(message.anchorTickerUuid)
	if err != nil {
		return err
	}

	// Store tags
	for _, tag := range message.tags {
		tagPrefix, err := TagKeyPrefix(tag)
		if err != nil {
			return err
		}
		err = streamStore.kvStore.Put(TagEntry(tagPrefix, newUuid), []byte(nil))
		if err != nil {
			return err
		}
	}

	// Store name
	namePrefix, err := NameKeyPrefix(message.name)
	if err != nil {
		return err
	}
	err = streamStore.kvStore.Put(NameEntry(namePrefix, newUuid), []byte(nil))
	if err != nil {
		return err
	}

	// Store previous node reference
	err = streamStore.kvStore.Put(PreviousNodeKey(newUuid), prevUuidBytes)
	if err != nil {
		return err
	}

	// Store anchor reference
	err = streamStore.kvStore.Put(AnchorNodeKey(newUuid), anchorUuidBytes)
	if err != nil {
		return err
	}

	// Store main record
	messageBytes, err := SerializeStreamImmutableMessage(message)
	err = streamStore.kvStore.Put(newUuid.String(), messageBytes)
	if err != nil {
		return err
	}

	// Set new head
	err = streamStore.setHead(message.subStreamID, message)
	if err != nil {
		return err
	}
	return nil
}

