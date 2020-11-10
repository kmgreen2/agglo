package ticker

import (
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
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

type StreamStore interface {
	GetMessagesByName(name string) ([]ImmutableMessage, error)
	GetMessagesByTags(tags []string) ([]ImmutableMessage, error)
	GetMessageByUUID(uuid gUuid.UUID) (ImmutableMessage, error)
	GetHistory(start gUuid.UUID, end gUuid.UUID) ([]ImmutableMessage, error)
	GetHistoryToLastAnchor(uuid gUuid.UUID) ([]ImmutableMessage, error)
	Append(message UncommittedMessage, subStreamID SubStreamID, anchorTickerUuid gUuid.UUID) (gUuid.UUID, error)

	// This function needs to be at a higher level
	// If the anchor is encoded into every message, then this is a simple fetch and compare from the Ticker: idx1 < idx2
	// In fact, that makes comparing cross-stream super easy
	//HappenedBefore(lhs *StreamImmutableMessage, rhs *StreamImmutableMessage) (bool, error)
}

type KVStreamStore struct {
	kvStore kvs.KVStore
	heads map[string]*StreamImmutableMessage
	digestType common.DigestType
	writeLocks map[string]*sync.Mutex
	streamLock *sync.Mutex
}

func (streamStore *KVStreamStore) GetMessagesByName(name string) ([]ImmutableMessage, error) {
	namePrefix, err := NameKeyPrefix(name)
	if err != nil {
		return nil, err
	}
	keys, err := streamStore.kvStore.List(namePrefix)
	if err != nil {
		return nil, err
	}
	messages := make([]ImmutableMessage, len(keys))
	for i, _ := range keys {
		messageBytes, err := streamStore.kvStore.Get(keys[i])
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

func (streamStore *KVStreamStore) GetMessagesByTags(tags []string) ([]ImmutableMessage, error) {
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

	messages := make([]ImmutableMessage, len(allKeys))
	for i, _ := range allKeys {
		messageBytes, err := streamStore.kvStore.Get(allKeys[i])
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

func (streamStore *KVStreamStore) GetMessageByUUID(uuid gUuid.UUID) (ImmutableMessage, error) {
	messageBytes, err := streamStore.kvStore.Get(uuid.String())
	if err != nil {
		return nil, err
	}
	return NewStreamImmutableMessageFromBuffer(messageBytes)
}

func (streamStore *KVStreamStore) GetMessages(uuids []gUuid.UUID) ([]ImmutableMessage, error) {
	var chainedMessages []ImmutableMessage
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

func (streamStore *KVStreamStore) GetHistory(start gUuid.UUID, end gUuid.UUID) ([]ImmutableMessage, error) {
	var chainedUuids []gUuid.UUID

	curr := end

	for {
		chainedUuids = append(chainedUuids, curr)
		if strings.Compare(curr.String(), start.String()) == 0 {
			break
		}
		prevBytes, err := streamStore.kvStore.Get(PreviousNodeKey(curr))
		if err != nil {
			return nil, err
		}
		prev, err := BytesToUUID(prevBytes)
		if err != nil {
			return nil, err
		}
		curr = prev
	}
	return streamStore.GetMessages(chainedUuids)
}

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

func (streamStore *KVStreamStore) GetHistoryToLastAnchor(uuid gUuid.UUID) ([]ImmutableMessage, error) {
	var chainedUuids []gUuid.UUID

	curr := uuid
	currAnchor, err := streamStore.getAnchorUUID(curr)
	if err != nil {
		return nil, err
	}
	chainedUuids = append(chainedUuids, curr)
	for {
		prevBytes, err := streamStore.kvStore.Get(PreviousNodeKey(curr))
		if err != nil {
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

func (streamStore *KVStreamStore) getHead(subStreamID SubStreamID) (*StreamImmutableMessage, error) {
	if head, ok := streamStore.heads[string(subStreamID)]; ok {
		return head , nil
	}
	return nil, common.NewNotFoundError(fmt.Sprintf("GetHead - cannot find substream: %s", subStreamID))
}

func (streamStore *KVStreamStore) setHead(subStreamID SubStreamID, message *StreamImmutableMessage) error {
	streamStore.heads[string(subStreamID)] = message
	return nil
}

// ToDo(KMG): Need to rollback any incomplete append
func (streamStore *KVStreamStore) Append(message UncommittedMessage, subStreamID SubStreamID,
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

	newUuid := immutableMessage.uuid
	prevUuidBytes, err := UuidToBytes(head.Uuid())
	if err != nil {
		return gUuid.Nil, err
	}

	anchorUuidBytes, err := UuidToBytes(anchorTickerUuid)
	if err != nil {
		return gUuid.Nil, err
	}

	// Store tags
	for _, tag := range immutableMessage.tags {
		tagPrefix, err := TagKeyPrefix(tag)
		if err != nil {
			return gUuid.Nil, err
		}
		err = streamStore.kvStore.Put(TagEntry(tagPrefix, newUuid), []byte(nil))
		if err != nil {
			return gUuid.Nil, err
		}
	}

	// Store name
	namePrefix, err := NameKeyPrefix(immutableMessage.name)
	if err != nil {
		return gUuid.Nil, err
	}
	err = streamStore.kvStore.Put(NameEntry(namePrefix, newUuid), []byte(nil))
	if err != nil {
		return gUuid.Nil, err
	}

	// Store previous node reference
	err = streamStore.kvStore.Put(PreviousNodeKey(newUuid), prevUuidBytes)
	if err != nil {
		return gUuid.Nil, err
	}

	// Store anchor reference
	err = streamStore.kvStore.Put(AnchorNodeKey(newUuid), anchorUuidBytes)
	if err != nil {
		return gUuid.Nil, err
	}

	// Store main record
	messageBytes, err := immutableMessage.Serialize()
	err = streamStore.kvStore.Put(newUuid.String(), messageBytes)
	if err != nil {
		return gUuid.Nil, err
	}

	err = streamStore.setHead(subStreamID, immutableMessage)
	if err != nil {
		return gUuid.Nil, err
	}

	return newUuid, nil
}

