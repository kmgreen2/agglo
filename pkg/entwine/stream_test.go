package entwine_test

import (
	gocrypto "crypto"
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKVStreamStore_GetMessagesByName(t *testing.T) {
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, uuidToName, err := test.GetKVStreamStore(6, entwine.SubStreamID("0"),
		messageSigner, gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for uuid, name := range uuidToName {
		messages, err := kvStreamStore.GetMessagesByName(name)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, messages, 1)
		assert.Equal(t, uuid, messages[0].Uuid().String())
	}
}

func TestKVStreamStore_GetMessagesByNameError(t *testing.T) {
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(6, entwine.SubStreamID("0"), messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = kvStreamStore.GetMessagesByName("notavalidname")
	assert.Error(t, err)
}

func TestKVStreamStore_GetMessagesByTags(t *testing.T) {
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, uuidToName, err := test.GetKVStreamStore(6, entwine.SubStreamID("0"),
		messageSigner, gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	tags := []string{"0", "3"}

	messages, err := kvStreamStore.GetMessagesByTags(tags)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, messages, 2)

	for i, message := range messages {
		assert.Equal(t, uuidToName[message.Uuid().String()], tags[i])
	}
}

func TestKVStreamStore_GetMessagesByTagsError(t *testing.T) {
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(6, entwine.SubStreamID("0"), messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = kvStreamStore.GetMessagesByTags([]string{"notavalidtag"})
	assert.Error(t, err)
}

func TestKVStreamStore_GetMessageByUUID(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(6, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	head, err := kvStreamStore.Head(subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	message, err := kvStreamStore.GetMessageByUUID(head.Uuid())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, head.Digest(), message.Digest())
}

func TestKVStreamStore_GetMessageByUUIDError(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(6, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = kvStreamStore.GetMessageByUUID(gUuid.New())
	assert.Error(t, err)
}

func TestKVStreamStore_GetMessages(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, uuidToName, err := test.GetKVStreamStore(6, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 0)
	for uuidStr, _ := range uuidToName {
		uuid, err := gUuid.Parse(uuidStr)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		uuids = append(uuids, uuid)
	}

	messages, err := kvStreamStore.GetMessages(uuids)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for _, message := range messages {
		name := uuidToName[message.Uuid().String()]
		assert.Equal(t, name, message.Name())
	}
}

func TestKVStreamStore_GetMessagesError(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(6, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 2)
	uuids[0] = gUuid.New()
	uuids[1] = gUuid.New()

	_, err = kvStreamStore.GetMessages(uuids)
	assert.Error(t, err)
}

func TestKVStreamStore_GetHistory(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, uuidToNames, err := test.GetKVStreamStore(12, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 12)

	for uuid, name := range uuidToNames {
		var index int
		fmt.Sscanf(name, "%d", &index)
		uuids[index] = gUuid.MustParse(uuid)
	}

	allMessages, err := kvStreamStore.GetMessages(uuids)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for i := 0; i < len(allMessages); i++ {
		for j := i; j < len(allMessages); j++ {
			historyMessages, err := kvStreamStore.GetHistory(uuids[i], uuids[j])
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			valid, err := entwine.ValidateStreamMessages(historyMessages, nil)
			assert.Nil(t, err)
			assert.True(t, valid)
		}
	}
}

func TestKVStreamStore_GetHistoryError(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(12, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = kvStreamStore.GetHistory(gUuid.New(), gUuid.New())
	assert.Error(t, err)
}

func TestKVStreamStore_GetHistoryToLastAnchor(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	kvStreamStore, _, _, uuidToNames, err := test.GetKVStreamStore(12, subStreamID, messageSigner,
		gUuid.New(), 4)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 12)

	for uuid, name := range uuidToNames {
		var index int
		fmt.Sscanf(name, "%d", &index)
		uuids[index] = gUuid.MustParse(uuid)
	}

	allMessages, err := kvStreamStore.GetMessages(uuids)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for i := 0; i < len(allMessages); i++ {
		historyMessages, err := kvStreamStore.GetHistoryToLastAnchor(uuids[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, (i % 4) + 2, len(historyMessages))
	}
}

func TestKVStreamStore_GetHistoryToLastAnchorError(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messageSigner, _, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	kvStreamStore, _, _, _, err := test.GetKVStreamStore(12, subStreamID, messageSigner,
		gUuid.New(), 0)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = kvStreamStore.GetHistoryToLastAnchor(gUuid.New())
	assert.Error(t, err)
}

// ToDo(KMG): Implement when we have ability to rollback
func TestKVStreamStore_AppendError(t *testing.T) {
}
