package entwine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"sort"
	"strings"
)

// MessageFingerprint is the fingerprint used to prove historical timelines of sub streams.  A fingerprint is
// derived from a StreamImmutableMessage
type MessageFingerprint struct {
	Signature []byte
	Digest []byte
	DigestType common.DigestType
	Uuid gUuid.UUID
	AnchorUuid gUuid.UUID
}

// Proof is the interface for proofs.  This is mostly here to make mocking possible,
// so we can easily test the proof functionality
type Proof interface {
	IsGenesis() (bool, error)
	Fingerprints() []*MessageFingerprint
	StartUuid() gUuid.UUID
	EndUuid() gUuid.UUID
	TickerUuid() gUuid.UUID
	SubStreamID() SubStreamID
	Validate() bool
	StartIdx() int64
	EndIdx() int64
	TickerIdx() int64
	String() string
	IsConsistent(prevProof Proof) (bool, error)
}

// ByAge implements sort.Interface for []Person based on
// the Age field.
type ByAnchor []Proof

func (a ByAnchor) Len() int           { return len(a) }
func (a ByAnchor) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByAnchor) Less(i, j int) bool {
	if a[i].SubStreamID().Equals(a[j].SubStreamID()) {
		return a[i].StartIdx() < a[j].EndIdx()
	}
	return a[i].TickerIdx() < a[j].TickerIdx()
}

// SortProofs will sort the provided proofs by their ticker index
func SortProofs(proofs []Proof) {
	sort.Sort(ByAnchor(proofs))
}

// Proof is used to encapsulate a sequence of immutable fingerprints for verification
type ProofImpl struct {
	messageFingerprints []*MessageFingerprint
	startUuid gUuid.UUID
	endUuid gUuid.UUID
	startIdx int64
	endIdx int64
	subStreamID SubStreamID
	tickerUuid gUuid.UUID
	tickerIdx int64
}

// String
func (proof *ProofImpl) String() string {
	return fmt.Sprintf("%s: [%d, %d] <-> %d", proof.SubStreamID(), proof.StartIdx(), proof.EndIdx(), proof.TickerIdx())
}

// Fingerprints will return the ordered fingerprints for this proof
func (proof *ProofImpl) Fingerprints() []*MessageFingerprint {
	return proof.messageFingerprints
}

// StartUuid will return the start UUID for this proof
func (proof *ProofImpl) StartUuid() gUuid.UUID {
	return proof.startUuid
}

// TickerIdx will return the idx of the ticker message for this proof
func (proof *ProofImpl) TickerIdx() int64 {
	return proof.tickerIdx
}

// StartIdx will return the idx of the start message for this proof
func (proof *ProofImpl) StartIdx() int64 {
	return proof.startIdx
}

// EndUuid will return the end UUID for this proof
func (proof *ProofImpl) EndUuid() gUuid.UUID {
	return proof.endUuid
}

// EndIdx will return the idx of the end message for this proof
func (proof *ProofImpl) EndIdx() int64 {
	return proof.endIdx
}

// TickerUuid will return the ticker UUID attached to this proof
func (proof *ProofImpl) TickerUuid() gUuid.UUID {
	return proof.tickerUuid
}

// SubstreamID will return the substream id for this proof
func (proof *ProofImpl) SubStreamID() SubStreamID {
	return proof.subStreamID
}

// SerializeProof serializes a Proof
func SerializeProof(proof *ProofImpl) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(proof.messageFingerprints)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.startUuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.endUuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.startIdx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.endIdx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.tickerIdx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.subStreamID)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(proof.tickerUuid)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeProof deserializes a Proof
func DeserializeProof(proofBytes []byte, proof *ProofImpl) error {
	byteBuffer := bytes.NewBuffer(proofBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&proof.messageFingerprints)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.startUuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.endUuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.startIdx)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.endIdx)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.tickerIdx)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.subStreamID)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&proof.tickerUuid)
	if err != nil {
		return err
	}
	return nil
}

// NewMessageFingerprint will create a MessageFingerprint from an ImmutableMessage
func NewMessageFingerprint(message *StreamImmutableMessage) *MessageFingerprint {
	return &MessageFingerprint{
		Signature: message.Signature(),
		Digest: message.Digest(),
		DigestType: message.DigestType(),
		Uuid: message.Uuid(),
		AnchorUuid: message.GetAnchorUUID(),
	}
}

// GenesisProofUuidBytes is the UUID used for all genesis proofs
var GenesisProofUuidBytes = []byte{220,241,234,178,108,41,73,73,162,150,124,204,66,118,33,160}

// NewGenesisProof will create a new genesis proof for a substream, anchored at the provided ticker message
func NewGenesisProof(subStreamID SubStreamID, message *TickerImmutableMessage) (*ProofImpl, error) {
	genesisProofUuid, err := gUuid.FromBytes(GenesisProofUuidBytes)
	if err != nil {
		return nil, err
	}
	return &ProofImpl {
		subStreamID: subStreamID,
		tickerUuid: message.Uuid(),
		startUuid: genesisProofUuid,
		endUuid: genesisProofUuid,
		startIdx: 0,
		endIdx: 0,
		tickerIdx: message.Index(),
	}, nil
}

// IsGenesis will return true if this proof is a genesis proof; false otherwise.  It may also return an error
// if something goes wrong.
func (proof *ProofImpl) IsGenesis() (bool, error) {
	startUuidBytes, err := proof.startUuid.MarshalBinary()
	if err != nil {
		return false, err
	}
	endUuidBytes, err := proof.startUuid.MarshalBinary()
	if err != nil {
		return false, err
	}
	return bytes.Compare(startUuidBytes, GenesisProofUuidBytes) == 0 && bytes.Compare(endUuidBytes,
		GenesisProofUuidBytes) == 0, nil
}

// NewProof will create a proof for a sequence of substream immutable messages, which are assumed to be anchored
// prior or with the same provided ticker message
//
func NewProof(messages []*StreamImmutableMessage, subStreamID SubStreamID,
	tickerMessage *TickerImmutableMessage) (*ProofImpl, error) {

	if len(messages) == 0 {
		msg := fmt.Sprintf("NewProof - no messages provided when creating proof for substreamID: %s", subStreamID)
		return nil, NewInvalidError(msg)
	}
	proof := &ProofImpl {
		messageFingerprints: make([]*MessageFingerprint, len(messages)),
		subStreamID: subStreamID,
		startUuid: messages[0].Uuid(),
		endUuid: messages[len(messages)-1].Uuid(),
		tickerUuid: tickerMessage.Uuid(),
		startIdx: messages[0].idx,
		endIdx: messages[len(messages)-1].idx,
		tickerIdx: tickerMessage.Index(),
	}

	for i, _ := range messages {
		proof.messageFingerprints[i] = NewMessageFingerprint(messages[i])
	}
	return proof, nil
}

// NewProofFromBytes will deserialize a byte slice into a proof; otherwise, return an error
func NewProofFromBytes(proofBytes []byte) (*ProofImpl, error) {
	proof := &ProofImpl{}
	err := DeserializeProof(proofBytes, proof)
	if err != nil {
		return nil, err
	}
	return proof, nil
}


// Validate will return true if the proof is valid; otherwise, return false
func (proof *ProofImpl) Validate() bool {
	var prevDigest []byte
	for i, fingerprint := range proof.messageFingerprints {
		if i > 0 {
			digest := common.InitHash(fingerprint.DigestType)
			digest.Write(prevDigest)
			digest.Write(fingerprint.Signature)
			if bytes.Compare(digest.Sum(nil), fingerprint.Digest) != 0 {
				return false
			}
		}
		prevDigest = fingerprint.Digest
	}
	return true
}

func isConsistent(lhs, rhs Proof) bool {
	// Both proofs must have at least one fingerprint
	if lhs == nil || rhs == nil || len(lhs.Fingerprints()) == 0 || len(rhs.Fingerprints()) == 0 {
		return false
	}

	// The last message of the lhs fingerprint must be the same as the first rhs fingerprint

	// First, check the UUIDs
	if strings.Compare(lhs.EndUuid().String(), rhs.StartUuid().String()) != 0 {
		return false
	}

	lhsFingerprints := lhs.Fingerprints()
	rhsFingerprints := rhs.Fingerprints()

	// Finally, verify the digests and signatures match
	lhsLastFingerprint := lhsFingerprints[len(lhsFingerprints)-1]
	rhsFirstFingerprint := rhsFingerprints[0]
	if bytes.Compare(lhsLastFingerprint.Digest, rhsFirstFingerprint.Digest) != 0 {
		return false
	}
	if bytes.Compare(lhsLastFingerprint.Signature, rhsFirstFingerprint.Signature) != 0 {
		return false
	}

	// Both proofs must be internally valid
	if !lhs.Validate() {
		return false
	}
	if !rhs.Validate() {
		return false
	}

	// All but the first anchor UUIDs in the rhs proof *must* be lhs.tickerUuid
	for i, fingerprint := range rhsFingerprints {
		if i == 0 {
			continue
		}
		if strings.Compare(lhs.TickerUuid().String(), fingerprint.AnchorUuid.String()) != 0 {
			return false
		}
	}

	// If the adjacent messages have the same fingerprints and both proofs are internally valid, then the proofs
	// contain a valid contiguous chain of fingerprints
	return true
}

// IsConsistent will return true if this proof is consistent with the previous proof; otherwise return false or an error
func (proof *ProofImpl) IsConsistent(prevProof Proof) (bool, error) {
	if prevProof == nil {
		return false, NewInvalidError(fmt.Sprintf("IsConsistent - nil previous proof provided"))
	}
	ok, err := prevProof.IsGenesis()
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return isConsistent(prevProof, proof), nil
}
