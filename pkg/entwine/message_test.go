package entwine_test

import (
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/serialization"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewStreamGenesisMessage(t *testing.T) {
	messages, authenticator, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 1,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	genesisMessage := messages[0]

	verified, err := genesisMessage.VerifySignature(authenticator)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.True(t, verified)
}

func TestNewStreamImmutableMessage(t *testing.T) {
	messages, authenticator, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 2,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	for i, message := range messages {
		verified, err := message.VerifySignature(authenticator)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.True(t, verified)

		if i > 0 {
			hashBytes, err := messages[i].ComputeChainHash(messages[i-1], authenticator)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, messages[i].Digest(), hashBytes)
		}
	}
}

func TestNewStreamImmutableMessageBadData(t *testing.T) {
	messages, _, objectStore, err := test.GetSubStream(entwine.SubStreamID("foobar"), 2,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	err = objectStore.Delete(messages[1].Name())
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = messages[1].Data()
	assert.Error(t, err)
}

func TestNewStreamImmutableMessageFromBuffer(t *testing.T) {
	newMessages := make([]*entwine.StreamImmutableMessage, 4)
	messages, authenticator, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 4,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for i, message := range messages {
		messageBytes, err := serialization.Serialize(message)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		newMessages[i], err = entwine.NewStreamImmutableMessageFromBuffer(messageBytes)
		if err != nil {
			assert.FailNow(t, err.Error())
		}

		if i > 0 {
			hashBytes, err := newMessages[i].ComputeChainHash(newMessages[i-1], authenticator)
			if err != nil {
				assert.FailNow(t, err.Error())
			}
			assert.Equal(t, messages[i].Digest(), hashBytes)
		}
	}
}

func TestNewStreamImmutableMessageFromBufferInvalid(t *testing.T) {
}

func TestStreamImmutableMessage_Serialize(t *testing.T) {
}

func TestStreamImmutableMessage_Data(t *testing.T) {
}

func TestStreamImmutableMessage_DataInvalid(t *testing.T) {
}

func TestStreamImmutableMessage_VerifySignature(t *testing.T) {
}

func TestStreamImmutableMessage_VerifySignatureInvalid(t *testing.T) {
}

func TestStreamImmutableMessage_GetSignaturePayload(t *testing.T) {
}

func TestStreamImmutableMessage_GetSignaturePayloadInvalid(t *testing.T) {
}

func TestStreamImmutableMessage_ComputeSignature(t *testing.T) {
}

func TestStreamImmutableMessage_ComputeSignatureInvalid(t *testing.T) {
}

func TestStreamImmutableMessage_ComputeChainHash(t *testing.T) {
}

func TestStreamImmutableMessage_ComputeChainHashInvalid(t *testing.T) {
}

func TestNewTickerImmutableMessage(t *testing.T) {
}

func TestNewTickerImmutableMessageBadSignature(t *testing.T) {
}

func TestNewTickerImmutableMessageFromBuffer(t *testing.T) {
}

func TestNewTickerImmutableMessageFromBufferInvalid(t *testing.T) {
}

func TestTickerImmutableMessage_Serialize(t *testing.T) {
}

func TestTickerImmutableMessage_SerializeInvalid(t *testing.T) {
}

func TestTickerImmutableMessage_ComputeSignature(t *testing.T) {
}

func TestTickerImmutableMessage_ComputeSignatureInvalid(t *testing.T) {
}

func TestTickerImmutableMessage_VerifySignature(t *testing.T) {
}

func TestTickerImmutableMessage_VerifySignatureInvalid(t *testing.T) {
}

func TestTickerImmutableMessage_GetSignaturePayload(t *testing.T) {
}

func TestTickerImmutableMessage_GetSignaturePayloadInvalid(t *testing.T) {
}

func TestTickerImmutableMessage_ComputeChainHash(t *testing.T) {
}

func TestTickerImmutableMessage_ComputeChainHashInvalid(t *testing.T) {
}

func TestComputeChainHash(t *testing.T) {
}

func TestComputeChainHashInvalid(t *testing.T) {
}

func TestValidateMessages(t *testing.T) {
}

func TestValidateMessagesInvalid(t *testing.T) {
}
