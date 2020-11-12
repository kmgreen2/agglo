package entwine

import (
	"bytes"
	"encoding/gob"
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

// NewProof will create a proof for  a sequence of substream immutable messages, which are assumed to be anchored
// at the provided ticker message
func NewProof(messages []*StreamImmutableMessage, subStreamID SubStreamID,
	tickerMessage *TickerImmutableMessage) *Proof {
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
	return proof
}

// NewProofFromBytes will deserialize a byte slice into a proof; otherwise, return an error
func NewProofFromBytes(proofBytes []byte) (*Proof, error) {
	proof := &Proof{}
	byteBuffer := bytes.NewBuffer(proofBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(proof)
	if err != nil {
		return nil, err
	}
	return proof, nil
}

// Validate will return true if the proof is valid; otherwise, return false
func (proof *Proof) Validate() bool {
	var prevDigest []byte
	for i, fingerprint := range proof.messageFingerprints {
		if i == 0 {
			prevDigest = fingerprint.Digest
		} else {
			digest := common.InitHash(fingerprint.DigestType)
			digest.Write(prevDigest)
			digest.Write(fingerprint.Signature)
			if bytes.Compare(digest.Sum(nil), fingerprint.Digest) != 0 {
				return false
			}
		}
	}
	return true
}

// Serialize will serialize the proof; otherwise return an error
func (proof *Proof) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(proof)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// IsConsistent will return true if this proof is consistent with the previous proof; otherwise return false or an error
func (proof *Proof) IsConsistent(prevProof *Proof) (bool, error) {
	ok, err := prevProof.IsGenesis()
	if err != nil {
		return false, err
	}
	if ok {
		return true, nil
	}
	return strings.Compare(prevProof.endUuid.String(), proof.startUuid.String()) == 0, nil
}
