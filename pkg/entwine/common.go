package entwine

import (
	"encoding/hex"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/util"
	"strings"
)

var invalidSequence []string = []string{":"}
func hasInvalidChars(id string) error {
	for _, sequence := range invalidSequence {
		if strings.Contains(id, sequence) {
			return NewInvalidError(fmt.Sprintf("ID (%s) contains invalid characters: %s", id, sequence))
		}
	}
	return nil
}

// PrimaryRecordKey returns the string representation of a primary record key from a UUID
func PrimaryRecordKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s:n", uuid.String())
}

// PreviousNodeKey returns the string representation of a previous record key from the previous record's UUID
func PreviousNodeKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s:p", uuid.String())
}

// AnchorNodeKey returns the string representation of a anchor node record key from the primary record's UUID
func AnchorNodeKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s:a", uuid.String())
}

// NameKeyPrefix will return the key prefix for a primary record's name
func NameKeyPrefix(name string) (string, error) {
	err := hasInvalidChars(name)
	if err != nil {
		return "", err
	}
	hasher := util.InitHash(util.MD5)
	_, err = hasher.Write([]byte(name))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:n:%s", hex.EncodeToString(hasher.Sum(nil))[:4], name), nil
}

// UuidFromNameKey will extract the Uuid from a name key
func UuidFromNameKey(key string) (string, error) {
	keySplit := strings.Split(key, ":")
	if len(keySplit) != 4 {
		return "", NewInvalidError(fmt.Sprintf("UuidFromNameKey - invalid name key: %s", key))
	}
	return keySplit[3], nil
}

// TagKeyPrefix will return the key prefix for a tag
func TagKeyPrefix(tag string) (string, error) {
	err := hasInvalidChars(tag)
	if err != nil {
		return "", err
	}
	hasher := util.InitHash(util.MD5)
	_, err = hasher.Write([]byte(tag))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:t:%s", hex.EncodeToString(hasher.Sum(nil))[:4], tag), nil
}

// UuidFromTagKey will extract the Uuid from a tag key
func UuidFromTagKey(key string) (string, error) {
	keySplit := strings.Split(key, ":")
	if len(keySplit) != 4 {
		return "", NewInvalidError(fmt.Sprintf("UuidFromTagKey - invalid tag key: %s", key))
	}
	return keySplit[3], nil
}

// TagEntry will return the string representation of a tag key from a prefix and primary record UUID
func TagEntry(tagPrefix string, uuid gUuid.UUID) string {
	return fmt.Sprintf("%s:%s", tagPrefix, uuid.String())
}

// NameEntry will return the string representation of a name key from a prefix and primary record UUID
func NameEntry(namePrefix string, uuid gUuid.UUID) string {
	return fmt.Sprintf("%s:%s", namePrefix, uuid.String())
}

// UuidToBytes converts a UUID to a byte slice and return an error if the UUID cannot be serialized
func UuidToBytes(uuid gUuid.UUID) ([]byte, error) {
	return uuid.MarshalBinary()
}

// BytesToUUID converts a byte slice into a UUID and return an error if the UUID cannot be deserialized
func BytesToUUID(uuidBytes []byte) (gUuid.UUID, error) {
	var newUuid gUuid.UUID
	err := newUuid.UnmarshalBinary(uuidBytes)
	if err != nil {
		return gUuid.Nil, err
	}
	return newUuid, nil
}

// ProofIdentifierPrefix will return the prefix of the representation of a proof entry and return an error if
// the prefix could not be derived.
func ProofIdentifierPrefix(subStreamID SubStreamID) (string, error) {
	err := hasInvalidChars(string(subStreamID))
	if err != nil {
		return "", err
	}
	hasher := util.InitHash(util.MD5)
	_, err = hasher.Write([]byte(subStreamID))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:pf", hex.EncodeToString(hasher.Sum(nil))[:4], subStreamID), nil
}

// ProofIdentifier will return the string representation of a proof entry key and return an error if
// the key could not be derived.
func ProofIdentifier(subStreamID SubStreamID, idx int) (string, error) {
	proofPrefix, err := ProofIdentifierPrefix(subStreamID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d", proofPrefix, idx), nil
}

// ReverseStreamMessages will reverse a list of stream immutable messages
func ReverseStreamMessages(messages []*StreamImmutableMessage) {
	if messages == nil {
		return
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
}

// ReverseTickerMessages will reverse a list of stream immutable messages
func ReverseTickerMessages(messages []*TickerImmutableMessage) {
	if messages == nil {
		return
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
}


