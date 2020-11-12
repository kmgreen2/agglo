package entwine_test

import (
	"encoding/hex"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/stretchr/testify/assert"
	"testing"
	"fmt"
)

func TestPrimaryRecordKey(t *testing.T) {
	uuid := gUuid.New()
	key := entwine.PrimaryRecordKey(uuid)
	assert.Equal(t, fmt.Sprintf("%s:n", uuid.String()), key)
}

func TestPreviousNodeKey(t *testing.T) {
	uuid := gUuid.New()
	key := entwine.PreviousNodeKey(uuid)
	assert.Equal(t, fmt.Sprintf("%s:p", uuid.String()), key)
}

func TestAnchorNodeKey(t *testing.T) {
	uuid := gUuid.New()
	key := entwine.AnchorNodeKey(uuid)
	assert.Equal(t, fmt.Sprintf("%s:a", uuid.String()), key)
}

func TestNameKeyPrefix(t *testing.T) {
	testName := "foobarbaz"
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(testName))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	namePrefix, err := entwine.NameKeyPrefix(testName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, fmt.Sprintf("%s:n:%s", hex.EncodeToString(hasher.Sum(nil))[:4], testName), namePrefix)
}

func TestNameEntry(t *testing.T) {
	uuid := gUuid.New()
	testName := "foobarbaz"
	namePrefix, err := entwine.NameKeyPrefix(testName)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, fmt.Sprintf("%s:%s", namePrefix, uuid.String()), entwine.NameEntry(namePrefix, uuid))
	extractedUuid, err := entwine.UuidFromNameKey(entwine.NameEntry(namePrefix, uuid))
	assert.Equal(t, uuid.String(), extractedUuid)
}

func TestUuidFromNameKeyInvalid(t *testing.T) {
	_, err := entwine.UuidFromNameKey("sdf:sdf-sdf0")
	assert.Error(t, err)
}

func TestUuidFromTagKeyInvalid(t *testing.T) {
	_, err := entwine.UuidFromTagKey("sdf:sdf-sdf0")
	assert.Error(t, err)
}

func TestTagEntry(t *testing.T) {
	uuid := gUuid.New()
	testTag := "tagfoobarbaz"
	tagPrefix, err := entwine.TagKeyPrefix(testTag)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, fmt.Sprintf("%s:%s", tagPrefix, uuid.String()), entwine.TagEntry(tagPrefix, uuid))
	extractedUuid, err := entwine.UuidFromTagKey(entwine.NameEntry(tagPrefix, uuid))
	assert.Equal(t, uuid.String(), extractedUuid)
}

func TestTagKeyPrefix(t *testing.T) {
	testTag := "foobarbaz"
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(testTag))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	tagPrefix, err := entwine.TagKeyPrefix(testTag)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, fmt.Sprintf("%s:t:%s", hex.EncodeToString(hasher.Sum(nil))[:4], testTag), tagPrefix)
}

func TestBytesToUUID(t *testing.T) {
	uuid := gUuid.New()

	uuidBytes, err := entwine.UuidToBytes(uuid)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	newUuid, err := entwine.BytesToUUID(uuidBytes)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, uuid.String(), newUuid.String())
}

func TestBytesToUUIDError(t *testing.T) {
	_, err := entwine.BytesToUUID([]byte{})
	assert.Error(t, err)
}

func TestProofIdentifierPrefix(t *testing.T) {
	subStreamID := entwine.SubStreamID("myssid")
	hasher := common.InitHash(common.MD5)
	_, err := hasher.Write([]byte(subStreamID))
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proofPrefix, err := entwine.ProofIdentifierPrefix(subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, fmt.Sprintf("%s:%s:pf", hex.EncodeToString(hasher.Sum(nil))[:4], subStreamID),
		proofPrefix)
}

func TestProofIdentifier(t *testing.T) {
	subStreamID := entwine.SubStreamID("myssid")
	proofPrefix, err := entwine.ProofIdentifierPrefix(subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proofID, err := entwine.ProofIdentifier(subStreamID, 1)

	assert.Equal(t, fmt.Sprintf("%s:%d", proofPrefix, 1), proofID)
}
