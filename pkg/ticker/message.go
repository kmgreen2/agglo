package ticker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/storage"
	"io"
)

// ImmutableMessage is the interface for all messages in a ticker or ticker substream
type ImmutableMessage interface {
	Signature() []byte
	Digest() []byte
	DigestType() common.DigestType
	Data() (io.Reader, error)
	Prev() uuid.UUID
	Uuid() uuid.UUID
	Name() string
	Tags() []string
	SubStream() SubStreamID
}

// BaseImmutableMessage is the base message for all streams
type BaseImmutableMessage struct {
	signature []byte
	digest []byte
	tickerDigest []byte
	digestType common.DigestType
	subStreamID SubStreamID
	uuid uuid.UUID
	prevUuid uuid.UUID
	ts int64
}

// StreamImmutableMessage is the message type for the main streams and sub-streams
type StreamImmutableMessage struct {
	BaseImmutableMessage
	name string
	tags []string
	objectDescriptor storage.ObjectDescriptor
	anchorTickerUuid uuid.UUID
}

func (message *StreamImmutableMessage) Signature() []byte {
	return message.signature
}

func (message *StreamImmutableMessage) Digest() []byte {
	return message.digest
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

func (message *StreamImmutableMessage) Prev() uuid.UUID {
	return message.prevUuid
}
func (message *StreamImmutableMessage) Uuid() uuid.UUID {
	return message.uuid
}
func (message *StreamImmutableMessage) Name() string {
	return message.name
}
func (message *StreamImmutableMessage) Tags() []string {
	return message.tags
}

// TickerImmutableMessage is the message type for the ticker stream
type TickerImmutableMessage struct {
	BaseImmutableMessage
	idx int64
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

func (message *TickerImmutableMessage) Prev() uuid.UUID {
	return message.prevUuid
}
func (message *TickerImmutableMessage) Uuid() uuid.UUID {
	return message.uuid
}
func (message *TickerImmutableMessage) Name() string {
	return fmt.Sprintf("%d", message.idx)
}
func (message *TickerImmutableMessage) Tags() []string {
	return make([]string, 0)
}

func (tickerMessage *TickerImmutableMessage) ComputeHash(prev *TickerImmutableMessage) ([]byte, error) {
	if prev.digestType != tickerMessage.digestType {
		return nil, NewInvalidError(fmt.Sprintf("ComputeHash - UUIDs (%s %s) mismatched digest types (%d %d)",
			prev.uuid.String(), tickerMessage.uuid.String(), prev.digestType, tickerMessage.digestType))
	}
	digest := common.InitHash(tickerMessage.digestType)
	byteBuffer := bytes.NewBuffer(make([]byte, 0))
	gEncoder := gob.NewEncoder(byteBuffer)
	err := gEncoder.Encode(tickerMessage.idx)
	if err != nil {
		return nil, err
	}
	err = gEncoder.Encode(tickerMessage.ts)
	if err != nil {
		return nil, err
	}
	uuidBytes, err := tickerMessage.uuid.MarshalBinary()
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(uuidBytes)
	if err != nil {
		return nil, err
	}
	_, err = byteBuffer.Write(tickerMessage.signature)
	if err != nil {
		return nil, err
	}
	_, err = digest.Write(append(prev.digest, byteBuffer.Bytes()...))
	if err != nil {
		return nil, err
	}
	return digest.Sum(nil), nil
}

func ComputeHash(lhs, rhs ImmutableMessage) ([]byte, error) {
	if lhsMessage, lhsOk := lhs.(*TickerImmutableMessage); lhsOk {
		if rhsMessage, rhsOk := rhs.(*TickerImmutableMessage); rhsOk {
			return rhsMessage.ComputeHash(lhsMessage)
		} else {
			return nil, NewInvalidError(fmt.Sprintf(
				"ComputeHash - mismatched message types for hashing UUIDs (%s %s)", lhs.Uuid(), rhs.Uuid()))
		}
	}
	return nil, nil
}

// ValidateMessages will validate a sequence of immutable messages
func ValidateMessages(messages []ImmutableMessage) (bool, error) {
	var prevMessage ImmutableMessage
	for i, message := range messages {
		if i == 0 {
			prevMessage = message
		} else {
			currDigest, err := ComputeHash(prevMessage, message)
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
