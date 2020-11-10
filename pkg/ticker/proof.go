package ticker

import (
	"bytes"
	"encoding/gob"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"strings"
)

type MessageFingerprint struct {
	signature []byte
	digest []byte
	digestType common.DigestType
	uuid gUuid.UUID
}

type Proof struct {
	messageFingerprints []*MessageFingerprint
	startUuid gUuid.UUID
	endUuid gUuid.UUID
	subStreamID SubStreamID
	tickerUuid gUuid.UUID
}

func NewMessageFingerprint(message ImmutableMessage) *MessageFingerprint {
	return &MessageFingerprint{
		signature: message.Signature(),
		digest: message.Digest(),
		digestType: message.DigestType(),
		uuid: message.Uuid(),
	}
}

func NewGenesisProof(subStreamID SubStreamID, message *TickerImmutableMessage) *Proof {
	return &Proof {
		subStreamID: subStreamID,
		tickerUuid: message.Uuid(),
	}
}

func NewProof(messages []ImmutableMessage, subStreamID SubStreamID, tickerMessage *TickerImmutableMessage) *Proof {
	proof := &Proof {
		messageFingerprints: make([]*MessageFingerprint, len(messages)),
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

func (proof *Proof) Validate() bool {
	var prevDigest []byte
	for i, fingerprint := range proof.messageFingerprints {
		if i == 0 {
			prevDigest = fingerprint.digest
		} else {
			digest := common.InitHash(fingerprint.digestType)
			digest.Write(prevDigest)
			digest.Write(fingerprint.signature)
			if bytes.Compare(digest.Sum(nil), fingerprint.digest) != 0 {
				return false
			}
		}
	}
	return true
}

func (proof *Proof) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(proof)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (proof *Proof) IsConsistent(prevProof *Proof) bool {
	return strings.Compare(prevProof.endUuid.String(), proof.startUuid.String()) == 0
}
