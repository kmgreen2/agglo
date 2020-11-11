package entwine

import (
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

	Primary record: <UUID>-n -> <data or data descriptor>, <substream id>, <object metadata>
	Previous node:  <UUID>-p -> <previous UUID>
	Proof: h(<substream_id>)[:4]-<substream-id>-pf-<idx> -> <serialized proof>
 */

// TickerStore is the interface for managing a ticker stream
type TickerStore interface {
	GetMessageByUUID(uuid gUuid.UUID) (ImmutableMessage, error)
	GetHistory(start gUuid.UUID, end gUuid.UUID) ([]ImmutableMessage, error)
	Anchor(proof []ImmutableMessage, subStreamID SubStreamID,
		authenticator crypto.Authenticator) (ImmutableMessage, error)
	HappenedBefore(lhs *TickerImmutableMessage, rhs *TickerImmutableMessage) (bool, error)
	Append(signer crypto.Signer) error
}

// KVStreamStore is an implementation of TickerStore that is backed by an in-memory map
type KVTickerStore struct {
	kvStore kvs.KVStore
	head *TickerImmutableMessage
	tickerLock *sync.Mutex
	proofIndexes map[string]int
	proofLocks map[string]*sync.Mutex
	digestType common.DigestType
}

// NewKVTickerStore returns a new KVStreamStore backed by the provided KVStore
// ToDo(KMG): Need to init heads, write locks from state in the backing KVStore
func NewKVTickerStore(kvStore kvs.KVStore, digestType common.DigestType) *KVTickerStore {
	return &KVTickerStore{
		kvStore: kvStore,
		digestType: digestType,
		proofIndexes: make(map[string]int),
		proofLocks: make(map[string]*sync.Mutex),
		tickerLock: &sync.Mutex{},
	}
}

// GetMessageByUUID will return the message with a given UUID; otherwise, return an error
func (tickerStore *KVTickerStore) GetMessageByUUID(uuid gUuid.UUID) (ImmutableMessage, error) {
	messageBytes, err := tickerStore.kvStore.Get(uuid.String())
	if err != nil {
		return nil, err
	}
	return NewTickerImmutableMessageFromBuffer(messageBytes)
}

// GetMessages will return the messages for the given UUIDs; otherwise, return an error
func (tickerStore *KVTickerStore) GetMessages(uuids []gUuid.UUID) ([]ImmutableMessage, error) {
	var chainedMessages []ImmutableMessage
	for _, myUuid := range uuids {
		messageBytes, err := tickerStore.kvStore.Get(myUuid.String())
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
func (tickerStore *KVTickerStore) GetHistory(start gUuid.UUID, end gUuid.UUID) ([]ImmutableMessage, error) {
	var chainedUuids []gUuid.UUID

	curr := end

	for {
		chainedUuids = append(chainedUuids, curr)
		if strings.Compare(curr.String(), start.String()) == 0 {
			break
		}
		prevBytes, err := tickerStore.kvStore.Get(PreviousNodeKey(curr))
		if err != nil {
			return nil, err
		}
		prev, err := BytesToUUID(prevBytes)
		if err != nil {
			return nil, err
		}
		curr = prev
	}
	return tickerStore.GetMessages(chainedUuids)
}

// CreateGenesisProof will create a genesis proof for a substream; otherwise, return an error
func (tickerStore *KVTickerStore) CreateGenesisProof(subStreamID SubStreamID) error {
	if _, ok := tickerStore.proofLocks[string(subStreamID)]; !ok {
		tickerStore.proofLocks[string(subStreamID)] = &sync.Mutex{}
	}
	tickerStore.proofLocks[string(subStreamID)].Lock()
	defer tickerStore.proofLocks[string(subStreamID)].Unlock()

	tickerMessage := tickerStore.head
	proof, err := NewGenesisProof(subStreamID, tickerMessage)
	if err != nil {
		return err
	}

	// Serialize and store the new proof
	proofBytes, err := SerializeProof(proof)
	if err != nil {
		return err
	}

	// ID == 0 because this is the genesis proof
	proofKey, err := ProofIdentifier(subStreamID, 0)
	if err != nil {
		return err
	}

	err = tickerStore.kvStore.Put(proofKey, proofBytes)
	if err != nil {
		return err
	}

	tickerStore.proofIndexes[string(subStreamID)] = 0

	return nil
}

// GetLatestProofKey will return the latest known proof for a substream; otherwise, return an error
func (tickerStore *KVTickerStore) GetLatestProofKey(subStreamID SubStreamID) (string, error) {
	proofPrefix, err := ProofIdentifierPrefix(subStreamID)
	if err != nil {
		return "", err
	}
	if idx, ok := tickerStore.proofIndexes[string(subStreamID)]; ok {
		return fmt.Sprintf("%s-%d", proofPrefix, idx), nil
	}


	keys, err := tickerStore.kvStore.List(proofPrefix)
	if err != nil {
		return "", err
	}

	// No proofs have been stored yet, return not found.  Caller should create a genesis proof
	if len(keys) == 0 {
		return "", common.NewNotFoundError(fmt.Sprintf(
			"GetProofStartUuid - cannot find previous proof for substream: %s", subStreamID))
	}

	maxIdx := -1

	for _, key := range keys {
		var idx int
		_, err = fmt.Sscanf(key, proofPrefix + "-%d", &idx)
		if err != nil {
			return "", err
		}
		if idx > maxIdx {
			maxIdx = idx
		}
	}
	tickerStore.proofIndexes[string(subStreamID)] = maxIdx
	return fmt.Sprintf("%s-%d", proofPrefix, maxIdx), nil
}

// GetProofStartUuid will return the UUID of the last message in the latest known proof for a substream; otherwise,
// return an error
func (tickerStore *KVTickerStore) GetProofStartUuid(subStreamID SubStreamID) (gUuid.UUID, error) {
	proofKey, err := tickerStore.GetLatestProofKey(subStreamID)
	if err != nil {
		return gUuid.Nil, err
	}
	proofBytes, err := tickerStore.kvStore.Get(proofKey)
	if err != nil {
		return gUuid.Nil, err
	}

	proof, err := NewProofFromBytes(proofBytes)
	if err != nil {
		return gUuid.Nil, err
	}

	return proof.endUuid, nil
}

// Anchor will return a ticker anchor, given a sequence of messages (used as a proof) for a substream; otherwise, return
// an error
func (tickerStore *KVTickerStore) Anchor(messages []ImmutableMessage, subStreamID SubStreamID,
	authenticator crypto.Authenticator) (ImmutableMessage, error) {

	tickerMessage := tickerStore.head
	// Create proof and validate
	proof := NewProof(messages, subStreamID, tickerMessage)
	if !proof.Validate() {
		return nil, NewInvalidError(fmt.Sprintf("Anchor - proof validation failed for substream: %s",
			subStreamID))
	}

	// Validate signatures
	for _, message := range messages {
		verified, err := message.VerifySignature(authenticator)
		if err != nil {
			return nil, err
		}
		if !verified {
			return nil, NewInvalidError(fmt.Sprintf("Anchor - invalid signature for uuid: %s", message.Uuid()))
		}
	}

	tickerStore.proofLocks[string(subStreamID)].Lock()
	defer tickerStore.proofLocks[string(subStreamID)].Unlock()

	// Get previous proof
	prevProofKey, err := tickerStore.GetLatestProofKey(subStreamID)
	if err != nil {
		return nil, err
	}

	prevBytes, err := tickerStore.kvStore.Get(prevProofKey)
	// There is no record of any previous proof, or something went wrong
	if err != nil {
		return nil, err
	}
	prevProof, err := NewProofFromBytes(prevBytes)
	if err != nil {
		return nil, err
	}

	// Validate the previous proof is consistent with the proposed proof
	ok, err := proof.IsConsistent(prevProof)
	if err != nil {
		return nil, NewInvalidError(fmt.Sprintf("Anchor - error with proposed proof: %s", err.Error()))
	}
	if !ok {
		return nil, NewInvalidError("Anchor - proposed proof is not consistent with previous chain of proof")
	}

	// Serialize and store the new proof
	proofBytes, err := SerializeProof(proof)
	if err != nil {
		return nil, err
	}

	proofKey, err := ProofIdentifier(subStreamID, tickerStore.proofIndexes[string(subStreamID)] + 1)
	if err != nil {
		return nil, err
	}

	err = tickerStore.kvStore.Put(proofKey, proofBytes)
	if err != nil {
		return nil, err
	}

	tickerStore.proofIndexes[string(subStreamID)]++

	return tickerMessage, nil
}

// HappenedBefore returns true if lhs happened before rhs; otherwise return an error and/or false
func (tickerStore *KVTickerStore) HappenedBefore(lhs *TickerImmutableMessage, rhs *TickerImmutableMessage) (bool,
	error) {
	return lhs.Index() < rhs.Index(), nil
}

// Append will append an uncommitted message to a substream; otherwise, return
// an error.
func (tickerStore *KVTickerStore) Append(signer crypto.Signer) error {
	ts := time.Now().Unix()
	tickerStore.tickerLock.Lock()
	defer tickerStore.tickerLock.Unlock()

	head := tickerStore.head

	immutableMessage, err := NewTickerImmutableMessage(tickerStore.digestType, signer,
		ts, head)
	if err != nil {
		return err
	}

	newUuid := immutableMessage.uuid

	// Store previous node reference
	prevUuidBytes, err := UuidToBytes(head.uuid)
	if err != nil {
		return err
	}
	err = tickerStore.kvStore.Put(PreviousNodeKey(newUuid), prevUuidBytes)
	if err != nil {
		return err
	}

	// Store main record
	messageBytes, err := SerializeTickerImmutableMessage(immutableMessage)
	err = tickerStore.kvStore.Put(newUuid.String(), messageBytes)
	if err != nil {
		return err
	}

	// Set new head
	tickerStore.head = immutableMessage

	return nil
}
