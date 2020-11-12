package entwine_test

import (
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKVTickerStore_GetMessageByUUID(t *testing.T) {
	var uuid gUuid.UUID
	tickerStore, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuid = tickerStore.Head().Uuid()
	for {
		message, err := tickerStore.GetMessageByUUID(uuid)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, uuid, message.Uuid())
		uuid = message.Prev()
		if uuid == gUuid.Nil {
			break
		}
	}
}

func TestKVTickerStore_GetMessageByUUIDError(t *testing.T) {
	tickerStore, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = tickerStore.GetMessageByUUID(gUuid.New())
	assert.Error(t, err)
}

func TestKVTickerStore_GetMessages(t *testing.T) {
	tickerStore, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 10)

	i := 0
	uuids[i] = tickerStore.Head().Uuid()
	for {
		message, err := tickerStore.GetMessageByUUID(uuids[i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		uuid := message.Prev()
		if uuid == gUuid.Nil {
			break
		}
		i++
		uuids[i] = uuid
	}

	for i := 1; i < 10; i++ {
		messages, err := tickerStore.GetMessages(uuids[:i])
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		for j, message := range messages {
			assert.Equal(t, uuids[j], message.Uuid())
		}
	}
}

func TestKVTickerStore_GetMessagesError(t *testing.T) {
	tickerStore, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = tickerStore.GetMessages([]gUuid.UUID{gUuid.New()})
	assert.Error(t, err)
}

func TestKVTickerStore_GetHistory(t *testing.T) {
	tickerStore, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	uuids := make([]gUuid.UUID, 10)

	uuids[9] = tickerStore.Head().Uuid()
	for i := 8; i >= 0; i-- {
		message, err := tickerStore.GetMessageByUUID(uuids[i+1])
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		uuid := message.Prev()
		if uuid == gUuid.Nil && i != 0 {
			assert.FailNow(t, "Encountered early end of chain")
		}
		uuids[i] = uuid
	}

	for i := 0; i < len(uuids); i++ {
		for j := i; j < len(uuids); j++ {
			historyMessages, err := tickerStore.GetHistory(uuids[i], uuids[j])
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			valid, err := entwine.ValidateTickerMessages(historyMessages, nil)
			assert.Nil(t, err)
			assert.True(t, valid)
		}
	}
}

func TestKVTickerStore_GetHistoryError(t *testing.T) {
}

func TestKVTickerStore_CreateGenesisProof(t *testing.T) {
}

func TestKVTickerStore_CreateGenesisProofError(t *testing.T) {
}

func TestKVTickerStore_GetLatestProofKey(t *testing.T) {
}

func TestKVTickerStore_GetLatestProofKeyError(t *testing.T) {
}

func TestKVTickerStore_GetProofStartUuid(t *testing.T) {
}

func TestKVTickerStore_GetProofStartUuidError(t *testing.T) {
}

func TestKVTickerStore_Anchor(t *testing.T) {
}

func TestKVTickerStore_AnchorError(t *testing.T) {
}

func TestKVTickerStore_HappenedBefore(t *testing.T) {
}

func TestKVTickerStore_Append(t *testing.T) {
}

