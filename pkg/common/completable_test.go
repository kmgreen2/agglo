package common

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompletableSuccess(t *testing.T) {
	completable := NewCompletable()

	err := completable.Success(int(1337))
	assert.Nil(t, err)

	val, err := completable.Future().Get()
	assert.Nil(t, err)
	assert.Equal(t, 1337, val.(int))
}

func TestCompletableFail(t *testing.T) {
	completable := NewCompletable()

	err := completable.Fail(NewInvalidError("Fail"))
	assert.Nil(t, err)

	_, err = completable.Future().Get()
	assert.Error(t, err)
}

func TestCompletableCancel(t *testing.T) {
	completable := NewCompletable()

	err := completable.Cancel()
	assert.Nil(t, err)

	_, err = completable.Future().Get()
	assert.Error(t, err)
}

