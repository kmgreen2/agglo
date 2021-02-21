package entwine_test

import (
	"github.com/golang/mock/gomock"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
	mocks "github.com/kmgreen2/agglo/test/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewProofFromBytesError(t *testing.T) {
	_, err := entwine.NewProofFromBytes([]byte{})
	assert.Error(t, err)
}

func TestProof_ValidateInvalid(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	invalidMessages, _, _, err := test.GetSubStream(subStreamID, 1, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(append(messages, invalidMessages...), subStreamID)

	assert.False(t, proof.Validate())
}

// This is tested by tickerStream.Anchor()
func TestProof_IsConsistent(t *testing.T) {
}

func TestProofImpl_HasUuid(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	for _, fingerprint := range proof.Fingerprints() {
		assert.True(t, proof.HasUuid(fingerprint.Uuid))
	}
	assert.False(t, proof.HasUuid(gUuid.New()))
}

func TestProof_IsConsistentInvalid(t *testing.T) {
	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	consistent, err := proof.IsConsistent(nil)
	assert.False(t, consistent)
	assert.Error(t, err)
}

// ToDo(KMG): Create an interface for Proofs, so these can be tested
func TestProof_IsConsistentNoFingerprints(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	prevProof := mocks.NewMockProof(ctrl)

	prevProof.EXPECT().
		Fingerprints().
		Return([]*entwine.MessageFingerprint{})

	prevProof.EXPECT().
		IsGenesis().
		Return(false, nil)

	consistent, err := proof.IsConsistent(prevProof)
	assert.False(t, consistent)
	assert.Nil(t, err)
}

func TestProof_IsConsistentMismatchAdjacent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	prevProof := mocks.NewMockProof(ctrl)

	prevProof.EXPECT().
		IsGenesis().
		Return(false, nil)

	prevProof.EXPECT().
		Fingerprints().
		Return([]*entwine.MessageFingerprint{entwine.NewMessageFingerprint(messages[0])})

	prevProof.EXPECT().
		EndUuid().
		Return(gUuid.New())

	consistent, err := proof.IsConsistent(prevProof)
	assert.False(t, consistent)
	assert.Nil(t, err)
}

func TestProof_IsConsistentMismatchInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	prevProof := mocks.NewMockProof(ctrl)

	prevProof.EXPECT().
		IsGenesis().
		Return(false, nil)

	prevProof.EXPECT().
		Fingerprints().
		Return([]*entwine.MessageFingerprint{entwine.NewMessageFingerprint(messages[1])}).
	    AnyTimes()

	prevProof.EXPECT().
		EndUuid().
		Return(messages[0].Uuid())

	consistent, err := proof.IsConsistent(prevProof)
	assert.False(t, consistent)
	assert.Nil(t, err)
}

func TestProof_IsConsistentMismatchLhsInvalid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	proof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	prevProof := mocks.NewMockProof(ctrl)

	prevProof.EXPECT().
		IsGenesis().
		Return(false, nil)

	prevProof.EXPECT().
		Fingerprints().
		Return([]*entwine.MessageFingerprint{entwine.NewMessageFingerprint(messages[0])}).
		AnyTimes()

	prevProof.EXPECT().
		EndUuid().
		Return(messages[0].Uuid())

	prevProof.EXPECT().
		Validate().
		Return(false)

	consistent, err := proof.IsConsistent(prevProof)
	assert.False(t, consistent)
	assert.Nil(t, err)
}

