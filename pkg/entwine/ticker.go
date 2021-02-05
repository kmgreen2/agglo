package entwine

import (
	"context"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"math"
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
	GetMessageByUUID(uuid gUuid.UUID) (*TickerImmutableMessage, error)
	GetHistory(start gUuid.UUID, end gUuid.UUID) ([]*TickerImmutableMessage, error)
	Anchor(proof []*StreamImmutableMessage, subStreamID SubStreamID,
		authenticator crypto.Authenticator) (*TickerImmutableMessage, error)
	HappenedBefore(lhs *TickerImmutableMessage, rhs *TickerImmutableMessage) (bool, error)
	Append(signer crypto.Signer) error
	GetLatestProofKey(subStreamID SubStreamID) (string, error)
	Head() *TickerImmutableMessage
	CreateGenesisProof(subStreamID SubStreamID) (Proof, error)
	GetProofStartUuid(subStreamID SubStreamID) (gUuid.UUID, error)
	GetProofForStreamIndex(subStreamID SubStreamID, streamIdx int64) (Proof, error)
	GetProofs(subStreamID SubStreamID, startIdx, endIdx int) ([]Proof, error)
}

// KVStreamStore is an implementation of TickerStore that is backed by an in-memory map
type KVTickerStore struct {
	kvStore      kvs.KVStore
	head         *TickerImmutableMessage
	tickerLock   *sync.Mutex
	proofIndexes map[string]int
	proofLocks   map[string]*sync.Mutex
	digestType   util.DigestType
}

// NewKVTickerStore returns a new KVStreamStore backed by the provided KVStore
// ToDo(KMG): Need to init heads, write locks from state in the backing KVStore
func NewKVTickerStore(kvStore kvs.KVStore, digestType util.DigestType) *KVTickerStore {
	return &KVTickerStore{
		kvStore: kvStore,
		digestType: digestType,
		proofIndexes: make(map[string]int),
		proofLocks: make(map[string]*sync.Mutex),
		tickerLock: &sync.Mutex{},
	}
}

// GetMessageByUUID will return the message with a given UUID; otherwise, return an error
func (tickerStore *KVTickerStore) GetMessageByUUID(uuid gUuid.UUID) (*TickerImmutableMessage, error) {
	messageBytes, err := tickerStore.kvStore.Get(context.Background(), uuid.String())
	if err != nil {
		return nil, err
	}
	return NewTickerImmutableMessageFromBuffer(messageBytes)
}

// GetMessages will return the messages for the given UUIDs; otherwise, return an error
func (tickerStore *KVTickerStore) GetMessages(uuids []gUuid.UUID) ([]*TickerImmutableMessage, error) {
	var chainedMessages []*TickerImmutableMessage
	for _, myUuid := range uuids {
		messageBytes, err := tickerStore.kvStore.Get(context.Background(), myUuid.String())
		if err != nil {
			return nil, err
		}
		message, err := NewTickerImmutableMessageFromBuffer(messageBytes)
		if err != nil {
			return nil, err
		}
		chainedMessages = append(chainedMessages, message)
	}
	return chainedMessages, nil
}

// Head will return the latest message appended to the ticker stream
func (tickerStore *KVTickerStore) Head() *TickerImmutableMessage {
	return tickerStore.head
}

// GetHistory will return the ordered, immutable history between two UUIDs; otherwise return an error
func (tickerStore *KVTickerStore) GetHistory(start gUuid.UUID, end gUuid.UUID) ([]*TickerImmutableMessage, error) {
	var chainedUuids []gUuid.UUID

	curr := end

	for {
		// If a nil start UUID is given, then process the entire history back to the genesis block
		if curr == gUuid.Nil && start == gUuid.Nil {
			break
		}

		if err := tickerStore.kvStore.Head(context.Background(), curr.String()); err != nil {
			return nil, err
		}
		chainedUuids = append(chainedUuids, curr)
		if strings.Compare(curr.String(), start.String()) == 0 {
			break
		}
		prevBytes, err := tickerStore.kvStore.Get(context.Background(), PreviousNodeKey(curr))
		// No previous message, assumes we have reached the first
		// ToDo(KMG): Do we care?  Should we check the first message and return an error?
		if errors.Is(err, &util.NotFoundError{}) {
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
	messages, err := tickerStore.GetMessages(chainedUuids)
	if err != nil {
		return nil, err
	}

	ReverseTickerMessages(messages)

	return messages, nil
}

// CreateGenesisProof will create a genesis proof for a substream; otherwise, return an error
func (tickerStore *KVTickerStore) CreateGenesisProof(subStreamID SubStreamID) (Proof, error) {
	if _, ok := tickerStore.proofLocks[string(subStreamID)]; !ok {
		tickerStore.proofLocks[string(subStreamID)] = &sync.Mutex{}
	}
	tickerStore.proofLocks[string(subStreamID)].Lock()
	defer tickerStore.proofLocks[string(subStreamID)].Unlock()

	tickerMessage := tickerStore.head
	proof, err := NewGenesisProof(subStreamID, tickerMessage)
	if err != nil {
		return nil, err
	}

	// Serialize and store the new proof
	proofBytes, err := SerializeProof(proof)
	if err != nil {
		return nil, err
	}

	// ID == 0 because this is the genesis proof
	proofKey, err := ProofIdentifier(subStreamID, 0)
	if err != nil {
		return nil, err
	}

	err = tickerStore.kvStore.Put(context.Background(), proofKey, proofBytes)
	if err != nil {
		return nil, err
	}

	tickerStore.proofIndexes[string(subStreamID)] = 0

	return proof, nil
}

// GetProofs will get all of the proofs at the given indexes for a sub stream
func (tickerStore *KVTickerStore) GetProofs(subStreamID SubStreamID, startIdx, endIdx int) ([]Proof, error) {
	var proofs []Proof
	if idx, ok := tickerStore.proofIndexes[string(subStreamID)]; ok {
		if startIdx > idx {
			return nil, NewInvalidError(fmt.Sprintf("GetProofs - start index is greater than current index: %d > %d",
				startIdx, idx))
		}
		if endIdx == -1 || endIdx > idx {
			endIdx = idx
		}
	} else {
		return nil, NewInvalidError(fmt.Sprintf("GetProofs - no proofs stored for '%s'",
			subStreamID))
	}

	for i := startIdx; i <= endIdx; i++ {
		proofID, err := ProofIdentifier(subStreamID, i)
		if err != nil {
			return nil, err
		}
		proofBytes, err := tickerStore.kvStore.Get(context.Background(), proofID)
		if err != nil {
			return nil, err
		}
		proof, err := NewProofFromBytes(proofBytes)
		if err != nil {
			return nil, err
		}
		proofs = append(proofs, proof)
	}
	return proofs, nil
}

// GetLatestProofKey will return the latest known proof for a substream; otherwise, return an error
func (tickerStore *KVTickerStore) GetLatestProofKey(subStreamID SubStreamID) (string, error) {
	proofPrefix, err := ProofIdentifierPrefix(subStreamID)
	if err != nil {
		return "", err
	}
	if idx, ok := tickerStore.proofIndexes[string(subStreamID)]; ok {
		return ProofIdentifier(subStreamID, idx)
	}

	keys, err := tickerStore.kvStore.List(context.Background(), proofPrefix)
	if err != nil {
		return "", err
	}

	// No proofs have been stored yet, return not found.  Caller should create a genesis proof
	if len(keys) == 0 {
		return "", util.NewNotFoundError(fmt.Sprintf(
			"GetProofStartUuid - cannot find previous proof for substream: %s", subStreamID))
	}

	maxIdx := -1

	for _, key := range keys {
		var idx int
		_, err = fmt.Sscanf(key, proofPrefix + ":%d", &idx)
		if err != nil {
			return "", err
		}
		if idx > maxIdx {
			maxIdx = idx
		}
	}

	// ToDo(KMG): This is opportunistic cache population.  Should this be validated instead of populated?
	tickerStore.proofLocks[string(subStreamID)].Lock()
	tickerStore.proofIndexes[string(subStreamID)] = maxIdx
	tickerStore.proofLocks[string(subStreamID)].Unlock()

	return ProofIdentifier(subStreamID, maxIdx)
}

// GetProofForStreamIndex returns the proof that contains the message at the provided index in the stream.  If no
// such proof can be found, it returns a NotFound error
func (tickerStore *KVTickerStore) GetProofForStreamIndex(subStreamID SubStreamID, streamIdx int64) (Proof, error) {
	if _, ok := tickerStore.proofIndexes[string(subStreamID)]; !ok {
		return nil, util.NewNotFoundError(fmt.Sprintf(
			"GetProofForStreamIndex - Cannot find proof for message at index '%d' for substream '%s'",
			streamIdx, subStreamID))
	}

	firstIdx := 0
	lastIdx := tickerStore.proofIndexes[string(subStreamID)]

	for firstIdx <= lastIdx {
		midIdx := int(math.Floor(float64(lastIdx + firstIdx) / 2))
		midProofID, err := ProofIdentifier(subStreamID, midIdx)
		if err != nil {
			return nil, err
		}
		midProofBytes, err := tickerStore.kvStore.Get(context.Background(), midProofID)
		if err != nil {
			return nil, err
		}
		midProof, err := NewProofFromBytes(midProofBytes)
		if err != nil {
			return nil, err
		}
		if streamIdx < midProof.StartIdx() {
			lastIdx = midIdx - 1
		} else if streamIdx > midProof.EndIdx() {
			firstIdx = midIdx + 1
		} else {
			return midProof, nil
		}
	}
	return nil, util.NewNotFoundError(fmt.Sprintf(
		"GetProofForStreamIndex - Cannot find proof for message at index '%d' for substream '%s'",
		streamIdx, subStreamID))
}

// GetProofStartUuid will return the UUID of the last message in the latest known proof for a substream; otherwise,
// return an error
// Note: Returns a Nil Uuid if there are no proofs beyond the genesis proof
func (tickerStore *KVTickerStore) GetProofStartUuid(subStreamID SubStreamID) (gUuid.UUID, error) {
	proofKey, err := tickerStore.GetLatestProofKey(subStreamID)
	if err != nil {
		return gUuid.Nil, err
	}
	proofBytes, err := tickerStore.kvStore.Get(context.Background(), proofKey)
	if err != nil {
		return gUuid.Nil, err
	}

	proof, err := NewProofFromBytes(proofBytes)
	if err != nil {
		return gUuid.Nil, err
	}

	isGenesis, err := proof.IsGenesis()
	if err != nil {
		return gUuid.Nil, err
	}
	if isGenesis {
		return gUuid.Nil, nil
	}

	return proof.endUuid, nil
}

// Anchor will return a ticker anchor, given a sequence of messages (used as a proof) for a substream; otherwise, return
// an error
// ToDo(KMG): Need to rollback if failure, then add tests for that
func (tickerStore *KVTickerStore) Anchor(messages []*StreamImmutableMessage, subStreamID SubStreamID,
	authenticator crypto.Authenticator) (*TickerImmutableMessage, error) {

	tickerMessage := tickerStore.head

	err := tickerStore.ValidateTickerUuids(messages, tickerMessage)
	if err != nil {
		return nil, err
	}

	// Create proof and validate
	proof, err := NewProof(messages, subStreamID, tickerMessage)
	if err != nil {
		return nil, err
	}
	// Create proof and validate
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

	prevBytes, err := tickerStore.kvStore.Get(context.Background(), prevProofKey)
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

	err = tickerStore.kvStore.Put(context.Background(), proofKey, proofBytes)
	if err != nil {
		return nil, err
	}

	tickerStore.proofIndexes[string(subStreamID)]++

	return tickerMessage, nil
}

// The following must hold (<message(s)>:tickerMessage)
// m[0]:t
// m[1:n]:t'
// tickerMessage
// t <= t' <= tickerMessage
func (tickerStore *KVTickerStore) ValidateTickerUuids(messages []*StreamImmutableMessage,
	currentTickerMessage *TickerImmutableMessage) error {
	firstUuid := messages[0].GetAnchorUUID()
	mainUuid := messages[1].GetAnchorUUID()

	// All but the first message *must* have the same anchor into the ticker
	for i := 2; i < len(messages); i++ {
		if strings.Compare(messages[i].GetAnchorUUID().String(), mainUuid.String()) != 0 {
			msg := fmt.Sprintf("ValidateTickerUuids - all anchor UUIDs, " +
				"except the first, must be the same to establish proof")
			return NewInvalidError(msg)
		}
	}

	firstAnchor, err := tickerStore.GetMessageByUUID(firstUuid)
	if err != nil {
		return err
	}
	mainAnchor, err := tickerStore.GetMessageByUUID(mainUuid)
	if err != nil {
		return err
	}

	firstBeforeMain, err := tickerStore.HappenedBefore(firstAnchor, mainAnchor)
	if err != nil {
		return err
	}

	mainBeforeCurrent, err := tickerStore.HappenedBefore(mainAnchor, currentTickerMessage)
	if err != nil {
		return err
	}

	if !firstBeforeMain && strings.Compare(firstUuid.String(), mainUuid.String()) != 0 {
		msg := fmt.Sprintf("ValidateTickerUuids - first message's anchor " +
			"did not happen before or same as main message anchor")
		return NewInvalidError(msg)
	}

	if !mainBeforeCurrent && strings.Compare(mainUuid.String(), currentTickerMessage.Uuid().String()) != 0 {
		msg := fmt.Sprintf("ValidateTickerUuids - main message anchor " +
			"did not happen before or same as current ticker message")
		return NewInvalidError(msg)
	}
	return nil
}

// HappenedBefore returns true if lhs happened before rhs; otherwise return an error and/or false
func (tickerStore *KVTickerStore) HappenedBefore(lhs *TickerImmutableMessage, rhs *TickerImmutableMessage) (bool,
	error) {
	return lhs.Index() < rhs.Index(), nil
}

// Append will append an uncommitted message to a sub stream; otherwise, return
// an error.
// ToDo(KMG): Need to rollback if failure, then add tests for that
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

	// Store previous node reference, if exists
	if head != nil {
		prevUuidBytes, err := UuidToBytes(head.uuid)
		if err != nil {
			return err
		}
		err = tickerStore.kvStore.Put(context.Background(), PreviousNodeKey(newUuid), prevUuidBytes)
		if err != nil {
			return err
		}
	}

	// Store main record
	messageBytes, err := SerializeTickerImmutableMessage(immutableMessage)
	err = tickerStore.kvStore.Put(context.Background(), newUuid.String(), messageBytes)
	if err != nil {
		return err
	}

	// Set new head
	tickerStore.head = immutableMessage

	return nil
}
