package entwine

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

// UncommittedMessage is a message provided by the caller to Append
type UncommittedMessage struct {
	objectDescriptor *storage.ObjectDescriptor
	name string
	tags []string
	signer crypto.Signer
}

func NewUncommittedMessage(objectDescriptor *storage.ObjectDescriptor, name string, tags []string,
	signer crypto.Signer) *UncommittedMessage {
	return &UncommittedMessage{
		objectDescriptor: objectDescriptor,
		name: name,
		tags: tags,
		signer: signer,
	}
}

// BaseImmutableMessage is the base message for all streams
type BaseImmutableMessage struct {
	signature   []byte
	digest      []byte
	digestType  common.DigestType
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
	subStreamID SubStreamID
}

// StreamImmutableMessage will return true if this message happened before other; false otherwise or if an error
// occurs.
func (message *StreamImmutableMessage) HappenedBefore(other *StreamImmutableMessage, tickerStore TickerStore) (bool,
	error) {
	if message.subStreamID.Equals(other.subStreamID) {
		return message.Index() < other.Index(), nil
	}

	proof, err := tickerStore.GetProofForStreamIndex(message.subStreamID, message.Index())
	if err != nil {
		return false, err
	}

	proofTickerAnchor, err := tickerStore.GetMessageByUUID(proof.TickerUuid())
	if err != nil {
		return false, err
	}

	otherTickerAnchor, err := tickerStore.GetMessageByUUID(other.GetAnchorUUID())
	if err != nil {
		return false, err
	}

	return proofTickerAnchor.Index() < otherTickerAnchor.Index(), nil
}

// SerializeStreamImmutableMessage serializes a StreamImmutableMessage
func SerializeStreamImmutableMessage(message *StreamImmutableMessage) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message.signature)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.digest)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.digestType)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.uuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.prevUuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.idx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.ts)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.name)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.tags)
	if err != nil {
		return nil, err
	}
	descBytes, err := storage.SerializeObjectDescriptor(message.objectDescriptor)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(descBytes)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.objectDigest)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.anchorTickerUuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.subStreamID)
	if err != nil {
		return nil, err
	}

	return byteBuffer.Bytes(), nil
}

// DeserializeStreamImmutableMessage deserializes a StreamImmutableMessage
func DeserializeStreamImmutableMessage(messageBytes []byte, message *StreamImmutableMessage) error {
	byteBuffer := bytes.NewBuffer(messageBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&message.signature)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.digest)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.digestType)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.uuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.prevUuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.idx)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.ts)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.name)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.tags)
	if err != nil {
		return err
	}
	var descBytes []byte
	err = gDecoder.Decode(&descBytes)
	if err != nil {
		return err
	}
	message.objectDescriptor = &storage.ObjectDescriptor{}
	err = storage.DeserializeObjectDescriptor(descBytes, message.objectDescriptor)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.objectDigest)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.anchorTickerUuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.subStreamID)
	if err != nil {
		return err
	}

	return nil
}

func NewStreamGenesisMessage(subStreamID SubStreamID, digestType common.DigestType,
	signer crypto.Signer, anchorTickerUuid gUuid.UUID) (*StreamImmutableMessage, error) {
	var err error
	message := &StreamImmutableMessage{}
	message.uuid = gUuid.New()
	message.subStreamID = subStreamID
	message.name = "genesis"
	message.digestType = digestType
	message.anchorTickerUuid = anchorTickerUuid
	message.idx = 0
	message.ts = 0
	message.objectDescriptor = storage.NewObjectDescriptor(&storage.NilObjectStoreBackendParams{}, "")
	message.objectDigest = []byte{}
	message.prevUuid = gUuid.Nil

	message.signature, err = message.ComputeSignature(signer)
	if err != nil {
		return nil, err
	}

	message.digest, err = message.ComputeChainHash(nil, nil)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// NewStreamImmutableMessage will create a new immutable message, which includes signing and hashing with the previous
// message in the chain
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

// NewStreamImmutableMessageFromBuffer will create a StreamImmutableMessage from a byte slice.  If it cannot
// decode the byte slice, an error will be returned
func NewStreamImmutableMessageFromBuffer(messageBytes []byte) (*StreamImmutableMessage, error) {
	message := &StreamImmutableMessage{}
	err := DeserializeStreamImmutableMessage(messageBytes, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Signature will return the signature computed when the message was created
func (message *StreamImmutableMessage) Signature() []byte {
	return message.signature
}

// GetAnchorUUID will return the anchor that connects this message to the ticker
func (message *StreamImmutableMessage) GetAnchorUUID() gUuid.UUID {
	return message.anchorTickerUuid
}

// Digest will return this message's digest, which is computed upon message creation
func (message *StreamImmutableMessage) Digest() []byte {
	return message.digest
}

// DataDigest will return the digest of the data backing the message, which is computed upon message creation
func (message *StreamImmutableMessage) DataDigest() []byte {
	return message.objectDigest
}

// SubStream will return the parent substream of this message
func (message *StreamImmutableMessage) SubStream() SubStreamID {
	return message.subStreamID
}

// DigestType will return the digest type used to compute the message and data digest
func (message *StreamImmutableMessage) DigestType() common.DigestType {
	return message.digestType
}

// Data will return a reader used to stream the data backing this message
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

// Prev will return the Uuid of the previous message
func (message *StreamImmutableMessage) Prev() gUuid.UUID {
	return message.prevUuid
}

// Uuid will return the Uuid of this message
func (message *StreamImmutableMessage) Uuid() gUuid.UUID {
	return message.uuid
}

// Name will return the caller-provided name of the message
func (message *StreamImmutableMessage) Name() string {
	return message.name
}

// Tags will return the tags associated with this message
func (message *StreamImmutableMessage) Tags() []string {
	return message.tags
}

// Index will return the substream index of this message
func (message *StreamImmutableMessage) Index() int64 {
	return message.idx
}

// VerifySignature will validate the signature of the message using a provided authenticator
func (message *StreamImmutableMessage) VerifySignature(authenticator crypto.Authenticator) (bool, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return false, err
	}
	signature, err := crypto.DeserializeSignature(message.signature)
	if err != nil {
		return false, err
	}
	return authenticator.Verify(signatureBytes, signature), nil
}

// GetSignaturePayload will return the serialized payload used to compute the signature of this message
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

// ComputeSignature will use the provided Signer to sign this message and return the resulting byte slice
func (message *StreamImmutableMessage) ComputeSignature(signer crypto.Signer) ([]byte, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign(signatureBytes)
	if err != nil {
		return nil, err
	}
	return crypto.SerializeSignature(signature)
}

// ComputeChainHash is a helper function that will compute this messages hash based on a provided previous node
// and authenticator.  The authenticator is used to verify the signature of this message, which will be skipped
// if the authenticator is nil.
func (message *StreamImmutableMessage) ComputeChainHash(prev *StreamImmutableMessage,
	authenticator crypto.Authenticator) ([]byte, error) {
	var prevDigest []byte
	var err error

	if prev != nil {
		if prev.digestType != message.digestType {
			return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - UUIDs (%s %s) mismatched digest types (%d %d)",
				prev.uuid.String(), message.uuid.String(), prev.digestType, message.digestType))
		}
		prevDigest = prev.digest
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
	if prevDigest != nil {
		_, err = digest.Write(append(prevDigest, message.signature...))
	} else {
		_, err = digest.Write(message.signature)
	}
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

// TickerImmutableMessage is the message type for the ticker stream
type TickerImmutableMessage struct {
	BaseImmutableMessage
}

// SerializeTickerImmutableMessage serializes a TickerImmutableMessage
func SerializeTickerImmutableMessage(message *TickerImmutableMessage) ([]byte, error) {
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(message.signature)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.digest)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.digestType)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.uuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.prevUuid)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.idx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(message.ts)
	if err != nil {
		return nil, err
	}
	return byteBuffer.Bytes(), nil
}

// DeserializeTickerImmutableMessage deserializes a TickerImmutableMessage
func DeserializeTickerImmutableMessage(messageBytes []byte, message *TickerImmutableMessage) error {
	byteBuffer := bytes.NewBuffer(messageBytes)
	gDecoder := gob.NewDecoder(byteBuffer)
	err := gDecoder.Decode(&message.signature)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.digest)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.digestType)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.uuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.prevUuid)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.idx)
	if err != nil {
		return err
	}
	err = gDecoder.Decode(&message.ts)
	if err != nil {
		return err
	}
	return nil
}

// NewTickerImmutableMessage will create a new immutable message, which includes signing and hashing with the previous
// message in the chain
func NewTickerImmutableMessage(digestType common.DigestType, signer crypto.Signer,
	ts int64, prevMessage *TickerImmutableMessage) (*TickerImmutableMessage, error) {
	var err error

	message := &TickerImmutableMessage{}
	message.digestType = digestType
	message.uuid = gUuid.New()
	if prevMessage != nil {
		message.prevUuid = prevMessage.Uuid()
		message.idx = prevMessage.idx + 1
	} else {
		message.prevUuid = gUuid.Nil
		message.idx = 0
	}
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

// NewTickerImmutableMessageFromBuffer will create a TickerImmutableMessage from a byte slice.  If it cannot
// decode the byte slice, an error will be returned
func NewTickerImmutableMessageFromBuffer(messageBytes []byte) (*TickerImmutableMessage, error) {
	message := &TickerImmutableMessage{}
	err := DeserializeTickerImmutableMessage(messageBytes, message)
	if err != nil {
		return nil, err
	}
	return message, nil
}

// Signature will return the signature computed when the message was created
func (message *TickerImmutableMessage) Signature() []byte {
	return message.signature
}

// Digest will return this message's digest, which is computed upon message creation
func (message *TickerImmutableMessage) Digest() []byte {
	return message.digest
}

// DigestType will return the digest type used to compute the message
func (message *TickerImmutableMessage) DigestType() common.DigestType {
	return message.digestType
}

// Data will return nil, since there is no data associated with a TickerImmutableMessage
func (message *TickerImmutableMessage) Data() (io.Reader, error) {
	return nil, nil
}

// DataDigest will return nil, since there is no data associated with a TickerImmutableMessage
func (message *TickerImmutableMessage) DataDigest() []byte {
	return nil
}

// Prev will return the Uuid of the previous message
func (message *TickerImmutableMessage) Prev() gUuid.UUID {
	return message.prevUuid
}

// Uuid will return the Uuid of this message
func (message *TickerImmutableMessage) Uuid() gUuid.UUID {
	return message.uuid
}

// Name returns the name of this message, which is the string representation of the message's index
func (message *TickerImmutableMessage) Name() string {
	return fmt.Sprintf("%d", message.idx)
}

// Tags returns an empty list, as there are no tags associated with a TickerImmutableMessage
func (message *TickerImmutableMessage) Tags() []string {
	return make([]string, 0)
}

// Index returns the index of this message in the ticker stream
func (message *TickerImmutableMessage) Index() int64 {
	return message.idx
}

// ComputeSignature will use the provided Signer to sign this message and return the resulting byte slice
func (message *TickerImmutableMessage) ComputeSignature(signer crypto.Signer) ([]byte, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return nil, err
	}
	signature, err := signer.Sign(signatureBytes)
	if err != nil {
		return nil, err
	}
	return crypto.SerializeSignature(signature)
}

// VerifySignature will use the provided Authenticator to verify this message
func (message *TickerImmutableMessage) VerifySignature(authenticator crypto.Authenticator) (bool, error) {
	signatureBytes, err := message.GetSignaturePayload()
	if err != nil {
		return false, err
	}
	signature, err := crypto.DeserializeSignature(message.signature)
	if err != nil {
		return false, err
	}
	return authenticator.Verify(signatureBytes, signature), nil
}

// GetSignaturePayload will return the serialized payload used to compute the signature of this message
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

// ComputeChainHash is a helper function that will compute this messages hash based on a provided previous node
// and authenticator.  The authenticator is used to verify the signature of this message, which will be skipped
// if the authenticator is nil.
func (message *TickerImmutableMessage) ComputeChainHash(prev *TickerImmutableMessage,
	authenticator crypto.Authenticator) ([]byte, error) {
	var prevDigest []byte
	var err error

	if prev != nil {
		if prev.digestType != message.digestType {
			return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - UUIDs (%s %s) mismatched digest types (%d %d)",
				prev.uuid.String(), message.uuid.String(), prev.digestType, message.digestType))
		}
		prevDigest = prev.digest
		if prev.digestType != message.digestType {
			return nil, NewInvalidError(fmt.Sprintf("ComputeChainHash - UUIDs (%s %s) mismatched digest types (%d %d)",
				prev.uuid.String(), message.uuid.String(), prev.digestType, message.digestType))
		}
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
	if prevDigest != nil {
		_, err = digest.Write(append(prevDigest, message.signature...))
	} else {
		_, err = digest.Write(message.signature)
	}
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

// ValidateStreamMessages will validate a sequence of immutable messages
func ValidateStreamMessages(messages []*StreamImmutableMessage, authenticator crypto.Authenticator) (bool, error) {
	var prevMessage *StreamImmutableMessage
	for i, message := range messages {
		if i > 0 {
			currDigest, err := message.ComputeChainHash(prevMessage, authenticator)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(message.Digest(), currDigest) {
				return false, nil
			}
		}
		prevMessage = message
	}
	return true, nil
}
// ValidateTickerMessages will validate a sequence of immutable messages
func ValidateTickerMessages(messages []*TickerImmutableMessage, authenticator crypto.Authenticator) (bool, error) {
	var prevMessage *TickerImmutableMessage
	for i, message := range messages {
		if i > 0 {
			currDigest, err := message.ComputeChainHash(prevMessage, authenticator)
			if err != nil {
				return false, err
			}
			if !bytes.Equal(message.Digest(), currDigest) {
				return false, nil
			}
		}
		prevMessage = message
	}
	return true, nil
}
