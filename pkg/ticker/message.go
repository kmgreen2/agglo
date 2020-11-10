package ticker

import (
	"bytes"
	"encoding/gob"
	"errors"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/crypto"
	"github.com/kmgreen2/agglo/pkg/storage"
	"io"
)

type UncommittedMessage struct {
	objectDescriptor *storage.ObjectDescriptor
	name string
	tags []string
	signer crypto.Signer
}

// ImmutableMessage is the interface for all messages in a ticker or ticker substream
type ImmutableMessage interface {
	Signature() []byte
	Digest() []byte
	DigestType() common.DigestType
	Data() (io.Reader, error)
	DataDigest() []byte
	Prev() gUuid.UUID
	Uuid() gUuid.UUID
	Name() string
	Tags() []string
	SubStream() SubStreamID
	Index() int64
	VerifySignature(authenticator crypto.Authenticator) (bool, error)

}

// BaseImmutableMessage is the base message for all streams
type BaseImmutableMessage struct {
	signature   []byte
	digest      []byte
	digestType  common.DigestType
	subStreamID SubStreamID
	uuid        gUuid.UUID
	prevUuid    gUuid.UUID
	idx         int64
	ts          int64
}

// StreamImmutableMessage is the message type for the main streams and sub-streams
type StreamImmutableMessage struct {
	BaseImmutableMessage
	name             string
	tags             []string
	objectDescriptor *storage.ObjectDescriptor
	objectDigest     []byte
	anchorTickerUuid gUuid.UUID
}

func NewStreamImmutableMessage(subStreamID SubStreamID, objectDescriptor *storage.ObjectDescriptor, name string,
	tags []string, digestType common.DigestType, signer crypto.Signer, ts int64, prevMessage *StreamImmutableMessage,
	anchorTickerUuid gUuid.UUID) (*StreamImmutableMessage, error) {
	message := &StreamImmutableMessage{}
	message.uuid = gUuid.New()
	message.subStreamID = subStreamID
	message.objectDescriptor = objectDescriptor
	message.name = name
	message.tags = tags
	message.digestType = digestType
	message.prevUuid = prevMessage.Uuid()
	message.anchorTickerUuid = anchorTickerUuid
	message.idx = prevMessage.idx + 1
	message.ts = ts

	// Compute data digest
	dataReader, err := message.Data()
	if err != nil {
		return nil, err
	}
	hasher := common.InitHash(digestType)
	byteBuf := make([]byte, 8192)
	for {
		n, err := dataReader.Read(byteBuf)
		if err != nil && errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		hasher.Write(byteBuf[:n])
	}
	message.objectDigest = hasher.Sum(nil)

	message.signature, err = message.ComputeSignature(signer)
	if err != nil {
		return nil, err
	}

	message.digest, err = message.ComputeChainHash(prevMessage, nil)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func NewStreamImmutableMessageFromBuffer(messageBytes []byte) (*StreamImmutableMessage, error) {
	message := &StreamImmutableMessage{}
	byteBuffer := bytes.NewBuffer(messageBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (message *StreamImmutableMessage) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (message *StreamImmutableMessage) Signature() []byte {
	return message.signature
}

func (message *StreamImmutableMessage) GetAnchorUUID() gUuid.UUID {
	return message.anchorTickerUuid
}

func (message *StreamImmutableMessage) Digest() []byte {
	return message.digest
}

func (message *StreamImmutableMessage) DataDigest() []byte {
	return message.objectDigest
}

func (message *StreamImmutableMessage) SubStream() SubStreamID {
	return message.subStreamID
}

func (message *StreamImmutableMessage) DigestType() common.DigestType {
	return message.digestType
}

func (message *StreamImmutableMessage) Data() (io.Reader, error) {
	objectStore, err := storage.NewObjectStore(message.objectDescriptor.GetParams())
	if err != nil {
		return nil, err
	}
	reader, err := objectStore.Get(message.objectDescriptor.GetKey())
	if err != nil {
		return nil, err
	}
	return reader, nil
}

func (message *StreamImmutableMessage) Prev() gUuid.UUID {
	return message.prevUuid
}

func (message *StreamImmutableMessage) Uuid() gUuid.UUID {
	return message.uuid
}

func (message *StreamImmutableMessage) Name() string {
	return message.name
}

func (message *StreamImmutableMessage) Tags() []string {
	return message.tags
}

func (message *StreamImmutableMessage) Index() int64 {
	return message.idx
}

func (message *StreamImmutableMessage) VerifySignature(authenticator crypto.Authenticator) (bool, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return false, err
	}
	signature, err := crypto.DeserializeSignature(message.signature)
	return authenticator.Verify(signatureBytes, signature), nil
}

func (message* StreamImmutableMessage) GetSignaturePayload() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message.name)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.subStreamID)
	if err != nil {
		return nil, err
	}
	anchorUuidBytes, err := message.anchorTickerUuid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(anchorUuidBytes)
	if err != nil {
		return nil, err
	}
	uuidBytes, err := message.uuid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(uuidBytes)
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(message.objectDigest)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (message *StreamImmutableMessage) ComputeSignature(signer crypto.Signer) ([]byte, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign(signatureBytes)
	if err != nil {
		return nil, err
	}
	return signature.Bytes(), nil
}

func (message *StreamImmutableMessage) ComputeChainHash(prev *StreamImmutableMessage,
	authenticator crypto.Authenticator) ([]byte, error) {
	if prev.digestType != message.digestType {
		return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - UUIDs (%s %s) mismatched digest types (%d %d)",
			prev.uuid.String(), message.uuid.String(), prev.digestType, message.digestType))
	}
	if message.signature == nil {
		return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - Cannot hash unsigned message UUID=%s",
			message.uuid.String()))
	}
	if authenticator != nil {
		verified, err := message.VerifySignature(authenticator)
		if err != nil {
			return nil, err
		}
		if !verified {
			return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - Invalid signature in message UUID=%s",
				message.uuid.String()))
		}
	}
	digest := common.InitHash(message.digestType)
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	_, err := byteBuffer.Write(message.signature)
	if err != nil {
		return nil, err
	}
	_, err = digest.Write(append(prev.digest, byteBuffer.Bytes()...))
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

// TickerImmutableMessage is the message type for the ticker stream
type TickerImmutableMessage struct {
	BaseImmutableMessage
}

func NewTickerImmutableMessage(subStreamID SubStreamID, digestType common.DigestType, signer crypto.Signer,
	ts int64, prevMessage *TickerImmutableMessage) (*TickerImmutableMessage, error) {
	var err error

	message := &TickerImmutableMessage{}
	message.subStreamID = subStreamID
	message.digestType = digestType
	message.prevUuid = prevMessage.Uuid()
	message.idx = prevMessage.idx + 1
	message.ts = ts

	message.signature, err = message.ComputeSignature(signer)
	if err != nil {
		return nil, err
	}

	message.digest, err = message.ComputeChainHash(prevMessage, nil)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func NewTickerImmutableMessageFromBuffer(messageBytes []byte) (*TickerImmutableMessage, error) {
	message := &TickerImmutableMessage{}
	byteBuffer := bytes.NewBuffer(messageBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

func (message *TickerImmutableMessage) Serialize() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (message *TickerImmutableMessage) Signature() []byte {
	return message.signature
}

func (message *TickerImmutableMessage) Digest() []byte {
	return message.digest
}

func (message *TickerImmutableMessage) SubStream() SubStreamID {
	return message.subStreamID
}

func (message *TickerImmutableMessage) DigestType() common.DigestType {
	return message.digestType
}

func (message *TickerImmutableMessage) Data() (io.Reader, error) {
	return nil, nil
}

func (message *TickerImmutableMessage) DataDigest() []byte {
	return nil
}

func (message *TickerImmutableMessage) Prev() gUuid.UUID {
	return message.prevUuid
}

func (message *TickerImmutableMessage) Uuid() gUuid.UUID {
	return message.uuid
}

func (message *TickerImmutableMessage) Name() string {
	return fmt.Sprintf("%d", message.idx)
}

func (message *TickerImmutableMessage) Tags() []string {
	return make([]string, 0)
}

func (message *TickerImmutableMessage) Index() int64 {
	return message.idx
}

func (message *TickerImmutableMessage) ComputeSignature(signer crypto.Signer) ([]byte, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign(signatureBytes)
	if err != nil {
		return nil, err
	}
	return signature.Bytes(), nil
}

func (message *TickerImmutableMessage) VerifySignature(authenticator crypto.Authenticator) (bool, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return false, err
	}
	signature, err := crypto.DeserializeSignature(message.signature)
	return authenticator.Verify(signatureBytes, signature), nil
}

func (message * TickerImmutableMessage) GetSignaturePayload() ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message.idx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.ts)
	if err != nil {
		return nil, err
	}
	uuidBytes, err := message.uuid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(uuidBytes)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

func (message *TickerImmutableMessage) ComputeChainHash(prev *TickerImmutableMessage,
	authenticator crypto.Authenticator) ([]byte, error) {
	if prev.digestType != message.digestType {
		return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - UUIDs (%s %s) mismatched digest types (%d %d)",
			prev.uuid.String(), message.uuid.String(), prev.digestType, message.digestType))
	}
	if message.signature == nil {
		return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - Cannot hash unsigned message UUID=%s",
			message.uuid.String()))
	}
	if authenticator != nil {
		verified, err := message.VerifySignature(authenticator)
		if err != nil {
			return nil, err
		}
		if !verified {
			return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - Invalid signature in message UUID=%s",
				message.uuid.String()))
		}
	}
	digest := common.InitHash(message.digestType)
	_, err := digest.Write(append(prev.digest, message.signature...))
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

func ComputeChainHash(lhs, rhs ImmutableMessage, authenticator crypto.Authenticator) ([]byte, error) {
	if lhsMessage, lhsOk := lhs.(*TickerImmutableMessage); lhsOk {
		if rhsMessage, rhsOk := rhs.(*TickerImmutableMessage); rhsOk {
			return rhsMessage.ComputeChainHash(lhsMessage, authenticator)
		} else {
			return nil, NewInvalidError(fmt.Sprintf(
				"ComputeChainHash - mismatched message types for hashing UUIDs (%s %s)", lhs.Uuid(), rhs.Uuid()))
		}
	} else if lhsMessage, lhsOk := lhs.(*StreamImmutableMessage); lhsOk {
		if rhsMessage, rhsOk := rhs.(*StreamImmutableMessage); rhsOk {
			return rhsMessage.ComputeChainHash(lhsMessage, authenticator)
		} else {
			return nil, NewInvalidError(fmt.Sprintf(
				"ComputeChainHash - mismatched message types for hashing UUIDs (%s %s)", lhs.Uuid(), rhs.Uuid()))
		}
	 }
	 return nil, NewInvalidError(fmt.Sprintf(
			"ComputeChainHash - invalid message types for hashing UUIDs (%s %s)", lhs.Uuid(), rhs.Uuid()))
}

// ValidateMessages will validate a sequence of immutable messages
func ValidateMessages(messages []ImmutableMessage, authenticator crypto.Authenticator) (bool, error) {
	var prevMessage ImmutableMessage
	for i, message := range messages {
		if i == 0 {
			prevMessage = message
		} else {
			currDigest, err := ComputeChainHash(prevMessage, message, authenticator)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(message.Digest(), currDigest) {
				return false, nil
			}
		}
	}
	return true, nil
}
