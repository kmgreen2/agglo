package entwine_test

import (
	"github.com/kmgreen2/agglo/pkg/entwine"
	"github.com/kmgreen2/agglo/test"
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
	subStreamID := entwine.SubStreamID("0")
	messages, _, _, err := test.GetSubStream(subStreamID, 4, false, nil)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	lhsProof, err := test.GetProof(messages, subStreamID)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	consistent, err := lhsProof.IsConsistent(nil)
	assert.False(t, consistent)
	assert.Error(t, err)
}

// ToDo(KMG): Create an interface for Proofs, so these can be tested
func TestProof_IsConsistentNoFingerprints(t *testing.T) {
}

func TestProof_IsConsistentMismatchAdjacent(t *testing.T) {
}

func TestProof_IsConsistentMismatchLhsInvalid(t *testing.T) {
}

func TestProof_IsConsistentMismatchRhsInvalid(t *testing.T) {
}

