package entwine_test

import (
	"context"
	gocrypto "crypto"
	gUuid "github.com/google/uuid"
	api "github.com/kmgreen2/agglo/generated/proto"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
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
		messageBytes, err := entwine.Serialize(message)
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
		messageBytes, err := entwine.Serialize(message)
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

func TestNewStreamImmutableMessageFromPb(t *testing.T) {
	pbMessage := &api.StreamImmutableMessage{
		Signature: []byte("signature"),
		Digest: []byte("digest"),
		DigestType: api.DigestType_SHA256,
		Uuid: gUuid.New().String(),
		PrevUuid: gUuid.New().String(),
		Idx: 1,
		Ts: time.Now().Unix(),
		Name: "foo",
		Tags: []string{"foo", "bar"},
		ObjectStoreConnectionString: "mem:foo",
		ObjectStoreKey: "foo",
		ObjectDigest: []byte("objdigest"),
		AnchorTickerUuid: gUuid.New().String(),
		SubStreamID: gUuid.New().String(),
	}

	message, err := entwine.NewStreamImmutableMessageFromPb(pbMessage)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, pbMessage.Signature, message.Signature())
	assert.Equal(t, pbMessage.Digest, message.Digest())
	assert.Equal(t, util.SHA256, message.DigestType())
	assert.Equal(t, pbMessage.Uuid, message.Uuid().String())
	assert.Equal(t, pbMessage.PrevUuid, message.Prev().String())
	assert.Equal(t, pbMessage.Idx, message.Index())
	assert.Equal(t, pbMessage.Ts, message.Ts())
	assert.Equal(t, pbMessage.Name, message.Name())
	assert.Equal(t, pbMessage.Tags, message.Tags())
	assert.Equal(t, pbMessage.AnchorTickerUuid, message.GetAnchorUUID().String())
	assert.Equal(t, pbMessage.SubStreamID, string(message.SubStream()))
}

func TestNewPbFromStreamImmutableMessage(t *testing.T) {
	origPbMessage := &api.StreamImmutableMessage{
		Signature: []byte("signature"),
		Digest: []byte("digest"),
		DigestType: api.DigestType_SHA256,
		Uuid: gUuid.New().String(),
		PrevUuid: gUuid.New().String(),
		Idx: 1,
		Ts: time.Now().Unix(),
		Name: "foo",
		Tags: []string{"foo", "bar"},
		ObjectStoreConnectionString: "mem:foo",
		ObjectStoreKey: "foo",
		ObjectDigest: []byte("objdigest"),
		AnchorTickerUuid: gUuid.New().String(),
		SubStreamID: gUuid.New().String(),
	}

	message, err := entwine.NewStreamImmutableMessageFromPb(origPbMessage)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	pbMessage := entwine.NewPbFromStreamImmutableMessage(message)

	assert.Equal(t, origPbMessage.Signature, pbMessage.Signature)
	assert.Equal(t, origPbMessage.Digest, pbMessage.Digest)
	assert.Equal(t, origPbMessage.DigestType, pbMessage.DigestType)
	assert.Equal(t, origPbMessage.Uuid, pbMessage.Uuid)
	assert.Equal(t, origPbMessage.PrevUuid, pbMessage.PrevUuid)
	assert.Equal(t, origPbMessage.Idx, pbMessage.Idx)
	assert.Equal(t, origPbMessage.Ts, pbMessage.Ts)
	assert.Equal(t, origPbMessage.Name, pbMessage.Name)
	assert.Equal(t, origPbMessage.Tags, pbMessage.Tags)
	assert.Equal(t, origPbMessage.AnchorTickerUuid, pbMessage.AnchorTickerUuid)
	assert.Equal(t, origPbMessage.SubStreamID, pbMessage.SubStreamID)
}

func TestNewTickerImmutableMessageFromPb(t *testing.T) {
	pbMessage := &api.TickerImmutableMessage{
		Signature: []byte("signature"),
		Digest: []byte("digest"),
		DigestType: api.DigestType_SHA256,
		Uuid: gUuid.New().String(),
		PrevUuid: gUuid.New().String(),
		Idx: 1,
		Ts: time.Now().Unix(),
	}

	message, err := entwine.NewTickerImmutableMessageFromPb(pbMessage)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, pbMessage.Signature, message.Signature())
	assert.Equal(t, pbMessage.Digest, message.Digest())
	assert.Equal(t, util.SHA256, message.DigestType())
	assert.Equal(t, pbMessage.Uuid, message.Uuid().String())
	assert.Equal(t, pbMessage.PrevUuid, message.Prev().String())
	assert.Equal(t, pbMessage.Idx, message.Index())
}

func TestNewPbFromTickerImmutableMessage(t *testing.T) {
	origPbMessage := &api.TickerImmutableMessage{
		Signature: []byte("signature"),
		Digest: []byte("digest"),
		DigestType: api.DigestType_SHA256,
		Uuid: gUuid.New().String(),
		PrevUuid: gUuid.New().String(),
		Idx: 1,
		Ts: time.Now().Unix(),
	}

	message, err := entwine.NewTickerImmutableMessageFromPb(origPbMessage)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	pbMessage := entwine.NewPbFromTickerImmutableMessage(message)

	assert.Equal(t, origPbMessage.Signature, pbMessage.Signature)
	assert.Equal(t, origPbMessage.Digest, pbMessage.Digest)
	assert.Equal(t, origPbMessage.DigestType, pbMessage.DigestType)
	assert.Equal(t, origPbMessage.Uuid, pbMessage.Uuid)
	assert.Equal(t, origPbMessage.PrevUuid, pbMessage.PrevUuid)
	assert.Equal(t, origPbMessage.Idx, pbMessage.Idx)
	assert.Equal(t, origPbMessage.Ts, pbMessage.Ts)
}
