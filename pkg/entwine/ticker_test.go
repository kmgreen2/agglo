package entwine_test

import (
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKVTickerStore_GetMessageByUUID(t *testing.T) {
	var uuid gUuid.UUID
	tickerStore, _, err := test.GetTickerStore(10)
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
	tickerStore, _, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = tickerStore.GetMessageByUUID(gUuid.New())
	assert.Error(t, err)
}

func TestKVTickerStore_GetMessages(t *testing.T) {
	tickerStore, _, err := test.GetTickerStore(10)
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
	tickerStore, _, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = tickerStore.GetMessages([]gUuid.UUID{gUuid.New()})
	assert.Error(t, err)
}

func TestKVTickerStore_GetHistory(t *testing.T) {
	tickerStore, _, err := test.GetTickerStore(10)
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
	tickerStore, _, err := test.GetTickerStore(10)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = tickerStore.GetHistory(gUuid.New(), gUuid.New())
	assert.Error(t, err)
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
	subStreamID := entwine.SubStreamID("0")
	startNumTicks := 12
	tickStride := 1
	messageStride := 3
	numEntanglements := 4
	tickerStore, streamStore, err := test.GetProofStream(startNumTicks, tickStride, messageStride, numEntanglements,
		subStreamID, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Get all ticker messages
	tickerHead := tickerStore.Head()
	tickerMessages, err := tickerStore.GetHistory(gUuid.Nil, tickerHead.Uuid())
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// Get all stream messages
	streamHead, err := streamStore.Head(subStreamID)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	streamMessages, err := streamStore.GetHistory(gUuid.Nil, streamHead.Uuid())
	if err != nil {
		assert.Fail(t, err.Error())
	}

	// Validate the following (assuming streams are created with GetProofStream(12, 1, 3, 4, subStreamID)
	//
	// TickerStream -> G (0) -> 1 -> 2 -> 3 -> ... -> 25
	// MessageStream -> G (0) -> 1 -> 2 -> 3 -> ... -> 12
	// Anchor:11 -> Messages:[0-3]
	// Anchor:14 -> Messages:[4-6]
	// Anchor:17 -> Messages:[7-9]
	// Anchor:20 -> Messages:[10-12]
	k := 0
	currTick := startNumTicks - 1
	for i := 0; i < numEntanglements; i++ {
		anchorUuid := tickerMessages[currTick].Uuid()
		// Genesis block must be accounted for
		if i == 0 {
			assert.Equal(t, anchorUuid.String(), streamMessages[0].GetAnchorUUID().String())
			k++
		}
		for j := 0; j < messageStride; j++ {
			assert.Equal(t, anchorUuid.String(), streamMessages[k].GetAnchorUUID().String())
			k++
			if k % tickStride == 0 {
				currTick++
			}
		}
	}
}

func TestKVTickerStore_GetProofForStreamIndex(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	startNumTicks := 12
	tickStride := 1
	messageStride := 3
	numEntanglements := 4
	tickerStore, streamStore, err := test.GetProofStream(startNumTicks, tickStride, messageStride, numEntanglements,
		subStreamID, true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Get all stream messages
	streamHead, err := streamStore.Head(subStreamID)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	streamMessages, err := streamStore.GetHistory(gUuid.Nil, streamHead.Uuid())
	if err != nil {
		assert.Fail(t, err.Error())
	}

	for i, message := range streamMessages {
		proof, err := tickerStore.GetProofForStreamIndex(subStreamID, message.Index())

		if i <= 9 {
			if err != nil {
				assert.Fail(t, err.Error())
			}
			assert.True(t, message.Index() <= proof.EndIdx())
			assert.True(t, message.Index() >= proof.StartIdx())
		} else {
			assert.True(t, errors.Is(err, &common.NotFoundError{}))
		}

	}
}

func TestKVTickerStore_GetProofs(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	startNumTicks := 12
	tickStride := 1
	messageStride := 3
	numEntanglements := 4
	tickerStore, streamStore, err := test.GetProofStream(startNumTicks, tickStride, messageStride, numEntanglements,
		subStreamID, true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	// Get all stream messages
	streamHead, err := streamStore.Head(subStreamID)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	_, err = streamStore.GetHistory(gUuid.Nil, streamHead.Uuid())
	if err != nil {
		assert.Fail(t, err.Error())
	}

	allProofs, err := tickerStore.GetProofs(subStreamID, 0, -1)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Len(t, allProofs, 4)

	allProofs, err = tickerStore.GetProofs(subStreamID, 0, 10)
	if err != nil {
		assert.Fail(t, err.Error())
	}
	assert.Len(t, allProofs, 4)

	for i := 0; i < 4; i++ {
		proofs, err := tickerStore.GetProofs(subStreamID, i, 4)
		if err != nil {
			assert.Fail(t, err.Error())
		}
		assert.Len(t, proofs, 4-i)
	}
}

func TestKVTickerStore_HappenedBefore(t *testing.T) {
}

func TestKVTickerStore_AnchorInvalidMessages(t *testing.T) {
}

func TestKVTickerStore_AnchorInvalidSignature(t *testing.T) {
}

// ToDo(KMG): Add these once we have rollback for Append and Anchor
func TestKVTickerStore_AnchorError(t *testing.T) {
}
func TestKVTickerStore_AppendError(t *testing.T) {
}

