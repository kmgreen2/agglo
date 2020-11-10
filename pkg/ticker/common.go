package ticker

import (
	"encoding/hex"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
)

func PrimaryRecordKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s-n", uuid.String())
}

func PreviousNodeKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s-p", uuid.String())
}

func AnchorNodeKey(uuid gUuid.UUID) string {
	return fmt.Sprintf("%s-a", uuid.String())
}

func NameKeyPrefix(name string) (string, error) {
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(name))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", hex.EncodeToString(hasher.Sum(nil))[:4], name), nil
}

func TagKeyPrefix(tag string) (string, error) {
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(tag))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", hex.EncodeToString(hasher.Sum(nil))[:4], tag), nil
}

func TagEntry(tagPrefix string, uuid gUuid.UUID) string {
	return fmt.Sprintf("%s-%s", tagPrefix, uuid.String())
}

func NameEntry(namePrefix string, uuid gUuid.UUID) string {
	return fmt.Sprintf("%s-%s", namePrefix, uuid.String())
}

func UuidToBytes(uuid gUuid.UUID) ([]byte, error) {
	return uuid.MarshalBinary()
}

func BytesToUUID(uuidBytes []byte) (gUuid.UUID, error) {
	var newUuid gUuid.UUID
	err := newUuid.UnmarshalBinary(uuidBytes)
	if err != nil {
		return gUuid.Nil, err
	}
	return newUuid, nil
}

func ProofIdentifierPrefix(subStreamID SubStreamID) (string, error) {
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(subStreamID))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s-pf", hex.EncodeToString(hasher.Sum(nil))[:4], subStreamID), nil
}

func ProofIdentifier(subStreamID SubStreamID, idx int) (string, error) {
	proofPrefix, err := ProofIdentifierPrefix(subStreamID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%d", proofPrefix, idx), nil
}

