package entwine_test

import (
	"context"
	gocrypto "crypto"
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
	messages, authenticator, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 4,
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

	err = objectStore.Delete(context.Background(), messages[1].Name())
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
	_, err := entwine.NewStreamImmutableMessageFromBuffer([]byte{})
	assert.Error(t, err)
}

func TestStreamImmutableMessage_Data(t *testing.T) {
	messages, _, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 2,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	reader, err := messages[1].Data()
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	dataBytes := make([]byte, 1024)
	n, err := reader.Read(dataBytes)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Equal(t, "0", string(dataBytes[:n]))
}

func TestStreamImmutableMessage_DataInvalid(t *testing.T) {
	messages, _, objectStore, err := test.GetSubStream(entwine.SubStreamID("foobar"), 2,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	objectStore.Delete(context.Background(), "0")

	_, err = messages[1].Data()
	assert.Error(t, err)
}

func TestStreamImmutableMessage_VerifySignatureInvalid(t *testing.T) {
	messages, _ , _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 2,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, wrongAuthenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verified, err := messages[0].VerifySignature(wrongAuthenticator)
	assert.Nil(t, err)
	assert.False(t, verified)
	verified, err = messages[1].VerifySignature(wrongAuthenticator)
	assert.Nil(t, err)
	assert.False(t, verified)
}

func TestNewTickerImmutableMessage(t *testing.T) {
	messages, authenticator, err := test.GetTickerStream(4)
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

func TestNewTickerImmutableMessageFromBuffer(t *testing.T) {
	newMessages := make([]*entwine.TickerImmutableMessage, 4)
	messages, authenticator, err := test.GetTickerStream(4)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for i, message := range messages {
		messageBytes, err := serialization.Serialize(message)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		newMessages[i], err = entwine.NewTickerImmutableMessageFromBuffer(messageBytes)
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

func TestNewTickerImmutableMessageFromBufferInvalid(t *testing.T) {
	_, err := entwine.NewTickerImmutableMessageFromBuffer([]byte{})
	assert.Error(t, err)
}

func TestTickerImmutableMessage_VerifySignatureInvalid(t *testing.T) {
	messages, _, err := test.GetTickerStream(2)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, wrongAuthenticator, _, err := test.GetSignerAuthenticator(gocrypto.SHA1)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verify, err := messages[0].VerifySignature(wrongAuthenticator)
	assert.Nil(t, err)
	assert.False(t, verify)
	verify, err = messages[1].VerifySignature(wrongAuthenticator)
	assert.Nil(t, err)
	assert.False(t, verify)

}

func TestValidateStreamMessages(t *testing.T) {
	messages, authenticator, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 4,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verify, err := entwine.ValidateStreamMessages(messages, authenticator)
	assert.Nil(t, err)
	assert.True(t, verify)
}

func TestValidateTickerMessages(t *testing.T) {
	messages, authenticator, err := test.GetTickerStream(4)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verify, err := entwine.ValidateTickerMessages(messages, authenticator)
	assert.Nil(t, err)
	assert.True(t, verify)
}

func TestValidateStreamMessagesInvalid(t *testing.T) {
	messagesBegin, _, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 4,
		false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	messagesEnd, _, _, err := test.GetSubStream(entwine.SubStreamID("foobar"), 4,
		false, nil)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verify, err := entwine.ValidateStreamMessages(append(messagesBegin, messagesEnd...), nil)
	assert.Nil(t, err)
	assert.False(t, verify)
}

func TestValidateTickerMessagesInvalid(t *testing.T) {
	messagesBegin, _, err := test.GetTickerStream(4)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	messagesEnd, _, err := test.GetTickerStream(4)

	if err != nil {
		assert.FailNow(t, err.Error())
	}

	verify, err := entwine.ValidateTickerMessages(append(messagesBegin, messagesEnd...), nil)
	assert.Nil(t, err)
	assert.False(t, verify)
}
