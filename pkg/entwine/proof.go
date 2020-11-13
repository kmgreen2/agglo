package entwine

import (
	"bytes"
	"encoding/gob"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"strings"
)

// messageFingerprint is the fingerprint used to prove historical timelines of substreams.  A fingerprint is
// derived from a StreamImmutableMessage
type messageFingerprint struct {
	Signature []byte
	Digest []byte
	DigestType common.DigestType
	Uuid gUuid.UUID
	AnchorUuid gUuid.UUID
}

// Proof is used to encapsulate a sequence of immutable fingerprints for verification
type Proof struct {
	messageFingerprints []*messageFingerprint
	startUuid gUuid.UUID
	endUuid gUuid.UUID
	subStreamID SubStreamID
	tickerUuid gUuid.UUID
}

// SerializeProof serializes a Proof
func SerializeProof(proof *Proof) ([]byte, error) {
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
func DeserializeProof(proofBytes []byte, proof *Proof) error {
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

// NewMessageFingerprint will create a messageFingerprint from an ImmutableMessage
func NewMessageFingerprint(message *StreamImmutableMessage) *messageFingerprint {
	return &messageFingerprint{
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
func NewGenesisProof(subStreamID SubStreamID, message *TickerImmutableMessage) (*Proof, error) {
	genesisProofUuid, err := gUuid.FromBytes(GenesisProofUuidBytes)
	if err != nil {
		return nil, err
	}
	return &Proof {
		subStreamID: subStreamID,
		tickerUuid: message.Uuid(),
		startUuid: genesisProofUuid,
		endUuid: genesisProofUuid,
	}, nil
}

// IsGenesis will return true if this proof is a genesis proof; false otherwise.  It may also return an error
// if something goes wrong.
func (proof *Proof) IsGenesis() (bool, error) {
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

// TickerUuid will return the ticker UUID associated with this proof
func (proof *Proof) TickerUuid() gUuid.UUID {
	return proof.tickerUuid
}

// NewProof will create a proof for a sequence of substream immutable messages, which are assumed to be anchored
// prior or with the same provided ticker message
//
func NewProof(messages []*StreamImmutableMessage, subStreamID SubStreamID,
	tickerMessage *TickerImmutableMessage) (*Proof, error) {

	if len(messages) == 0 {
		msg := fmt.Sprintf("NewProof - no messages provided when creating proof for substreamID: %s", subStreamID)
		return nil, NewInvalidError(msg)
	}
	proof := &Proof {
		messageFingerprints: make([]*messageFingerprint, len(messages)),
		subStreamID: subStreamID,
		startUuid: messages[0].Uuid(),
		endUuid: messages[len(messages)-1].Uuid(),
		tickerUuid: tickerMessage.Uuid(),
	}

	for i, _ := range messages {
		proof.messageFingerprints[i] = NewMessageFingerprint(messages[i])
	}
	return proof, nil
}

// NewProofFromBytes will deserialize a byte slice into a proof; otherwise, return an error
func NewProofFromBytes(proofBytes []byte) (*Proof, error) {
	proof := &Proof{}
	err := DeserializeProof(proofBytes, proof)
	if err != nil {
		return nil, err
	}
	return proof, nil
}


// Validate will return true if the proof is valid; otherwise, return false
func (proof *Proof) Validate() bool {
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

func isConsistent(lhs, rhs *Proof) bool {
	// Both proofs must have at least one fingerprint
	if lhs == nil || rhs == nil || len(lhs.messageFingerprints) == 0 || len(rhs.messageFingerprints) == 0 {
		return false
	}

	// The last message of the lhs fingerprint must be the same as the first rhs fingerprint

	// First, check the UUIDs
	if strings.Compare(lhs.endUuid.String(), rhs.startUuid.String()) != 0 {
		return false
	}

	// Finally, verify the digests and signatures match
	lhsLastFingerprint := lhs.messageFingerprints[len(lhs.messageFingerprints)-1]
	rhsFirstFingerprint := rhs.messageFingerprints[0]
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
	for i, message := range rhs.messageFingerprints {
		if i == 0 {
			continue
		}
		if strings.Compare(lhs.TickerUuid().String(), message.AnchorUuid.String()) != 0 {
			return false
		}
	}

	// If the adjacent messages have the same fingerprints and both proofs are internally valid, then the proofs
	// contain a valid contiguous chain of fingerprints
	return true
}

// IsConsistent will return true if this proof is consistent with the previous proof; otherwise return false or an error
func (proof *Proof) IsConsistent(prevProof *Proof) (bool, error) {
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
