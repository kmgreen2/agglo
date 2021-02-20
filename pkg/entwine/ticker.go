package entwine

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/state"
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
	Head() (*TickerImmutableMessage, error)
	CreateGenesisProof(subStreamID SubStreamID) (Proof, error)
	GetProofStartUuid(subStreamID SubStreamID) (gUuid.UUID, error)
	GetProofForStreamIndex(subStreamID SubStreamID, streamIdx int64) (Proof, error)
	GetProofs(subStreamID SubStreamID, startIdx, endIdx int) ([]Proof, error)
	GetLatestProof(subStreamID SubStreamID) (Proof, error)
}

// KVStreamStore is an implementation of TickerStore that is backed by an in-memory map
type KVTickerStore struct {
	kvStore      kvs.KVStore
	tickerLock   *sync.Mutex
	proofLocks   map[string]*state.KVDistributedLock
	digestType   util.DigestType
}

// NewKVTickerStore returns a new KVStreamStore backed by the provided KVStore
// ToDo(KMG): Need to init heads, write locks from state in the backing KVStore
func NewKVTickerStore(kvStore kvs.KVStore, digestType util.DigestType) *KVTickerStore {
	return &KVTickerStore{
		kvStore: kvStore,
		digestType: digestType,
		proofLocks: make(map[string]*state.KVDistributedLock),
		tickerLock: &sync.Mutex{},
	}
}

func (tickerStore *KVTickerStore) getLatestProofIndex(id SubStreamID) (int, error) {
	var idx int
	idxBytes, err := tickerStore.kvStore.Get(context.Background(), ProofIndexKey(id))
	if err != nil {
		return -1, err
	}
	currDecoder := gob.NewDecoder(bytes.NewBuffer(idxBytes))
	err = currDecoder.Decode(&idx)
	if err != nil {
		return -1, err
	}
	return idx, nil
}

func (tickerStore *KVTickerStore) setLatestProofIndex(id SubStreamID, idx int) error {
	// Index must monotonically increase by 1 as each proof is added, so we can explicitly enforce
	prevBuffer := bytes.NewBuffer([]byte{})
	prevEncoder := gob.NewEncoder(prevBuffer)
	err := prevEncoder.Encode(idx-1)
	if err != nil {
		return err
	}
	currBuffer := bytes.NewBuffer([]byte{})
	currEncoder := gob.NewEncoder(currBuffer)
	err = currEncoder.Encode(idx)
	if err != nil {
		return err
	}
	if idx > 0 {
		return tickerStore.kvStore.AtomicPut(context.Background(), ProofIndexKey(id), prevBuffer.Bytes(),
			currBuffer.Bytes())
	} else {
		return tickerStore.kvStore.AtomicPut(context.Background(), ProofIndexKey(id), nil,
			currBuffer.Bytes())
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
func (tickerStore *KVTickerStore) Head() (*TickerImmutableMessage, error) {
	messageUUIDBytes, err := tickerStore.kvStore.Get(context.Background(), TickerHeadKey())
	if err != nil {
		return nil, err
	}

	messageUUID, err := BytesToUUID(messageUUIDBytes)
	if err != nil {
		return nil, err
	}

	message, err := tickerStore.GetMessageByUUID(messageUUID)
	if err != nil {
		return nil, err
	}
	return message, nil
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

func (tickerStore *KVTickerStore) getProofLock(subStreamID SubStreamID) *state.KVDistributedLock {
	if _, ok := tickerStore.proofLocks[string(subStreamID)]; !ok {
		tickerStore.proofLocks[string(subStreamID)] = state.NewKVDistributedLock(string(subStreamID),
			tickerStore.kvStore)
	}
	return tickerStore.proofLocks[string(subStreamID)]
}

// CreateGenesisProof will create a genesis proof for a substream; otherwise, return an error
func (tickerStore *KVTickerStore) CreateGenesisProof(subStreamID SubStreamID) (Proof, error) {
	proofLock := tickerStore.getProofLock(subStreamID)
	// Take lock
	ctx, err := proofLock.Lock(context.Background(), -1)
	if err != nil {
		return nil, err
	}
	defer func() {
		// ToDo(KMG): Log the error
		_ = proofLock.Unlock(ctx)
	}()

	tickerMessage, err := tickerStore.Head()
	if err != nil {
		return nil, err
	}
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

	err = tickerStore.setLatestProofIndex(subStreamID, 0)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// GetProofs will get all of the proofs at the given indexes for a sub stream
func (tickerStore *KVTickerStore) GetProofs(subStreamID SubStreamID, startIdx, endIdx int) ([]Proof, error) {
	var proofs []Proof
	if idx, err := tickerStore.getLatestProofIndex(subStreamID); err == nil {
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
	if idx, err := tickerStore.getLatestProofIndex(subStreamID); err == nil {
		return ProofIdentifier(subStreamID, idx)
	} else {
		return "", err
	}
}

// GetLatestProof will return the latest proof for a given substream
func (tickerStore *KVTickerStore) GetLatestProof(subStreamID SubStreamID) (Proof, error) {
	proofKey, err := tickerStore.GetLatestProofKey(subStreamID)
	if err != nil {
		return nil, err
	}
	proofBytes, err := tickerStore.kvStore.Get(context.Background(), proofKey)
	if err != nil {
		return nil, err
	}
	proof, err := NewProofFromBytes(proofBytes)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// GetProofForStreamIndex returns the proof that contains the message at the provided index in the stream.  If no
// such proof can be found, it returns a NotFound error
func (tickerStore *KVTickerStore) GetProofForStreamIndex(subStreamID SubStreamID, streamIdx int64) (Proof, error) {
	firstIdx := 0
	lastIdx := 0

	if idx, err := tickerStore.getLatestProofIndex(subStreamID); err != nil {
		return nil, util.NewNotFoundError(fmt.Sprintf(
			"GetProofForStreamIndex - Cannot find proof for message at index '%d' for substream '%s'",
			streamIdx, subStreamID))
	} else {
		lastIdx = idx
	}

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

	tickerMessage, err := tickerStore.Head()
	if err != nil {
		return nil, err
	}
	err = tickerStore.ValidateTickerUuids(messages, tickerMessage)
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

	proofLock := tickerStore.getProofLock(subStreamID)
	// Take lock
	ctx, err := proofLock.Lock(context.Background(), -1)
	if err != nil {
		return nil, err
	}
	defer func() {
		// ToDo(KMG): Log the error
		_ = proofLock.Unlock(ctx)
	}()

	latestProofIndex, err := tickerStore.getLatestProofIndex(subStreamID)
	if err != nil {
		return nil, err
	}

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

	proofKey, err := ProofIdentifier(subStreamID, latestProofIndex + 1)
	if err != nil {
		return nil, err
	}

	err = tickerStore.kvStore.Put(context.Background(), proofKey, proofBytes)
	if err != nil {
		return nil, err
	}

	err = tickerStore.setLatestProofIndex(subStreamID, latestProofIndex + 1)
	if err != nil {
		return nil, err
	}

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

// setHead will set the head of the ticker; otherwise, return an error
func (tickerStore *KVTickerStore) setHead(message *TickerImmutableMessage) error {
	var err error
	var prevUUIDBytes  []byte
	messageUUIDBytes, err := UuidToBytes(message.uuid)
	if err != nil {
		return err
	}
	if message.prevUuid != gUuid.Nil {
		prevUUIDBytes, err = UuidToBytes(message.prevUuid)
		if err != nil {
			return err
		}
	}
	return tickerStore.kvStore.AtomicPut(context.Background(), TickerHeadKey(), prevUUIDBytes,
		messageUUIDBytes)
}

// Append will append an uncommitted message to a sub stream; otherwise, return
// an error.
// ToDo(KMG): Need to rollback if failure, then add tests for that
func (tickerStore *KVTickerStore) Append(signer crypto.Signer) error {
	ts := time.Now().Unix()
	tickerStore.tickerLock.Lock()
	defer tickerStore.tickerLock.Unlock()

	head, err := tickerStore.Head()
	if err != nil && !errors.Is(err, &util.NotFoundError{}) {
		return err
	}

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
	return tickerStore.setHead(immutableMessage)
}
