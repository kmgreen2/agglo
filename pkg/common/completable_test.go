package common_test

import (
	"github.com/kmgreen2/agglo/pkg/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompletableSuccess(t *testing.T) {
	completable := common.NewCompletable()

	err := completable.Success(int(1337))
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 1337, result.Value().(int))
}

func TestCompletableFail(t *testing.T) {
	completable := common.NewCompletable()

	err := completable.Fail(common.NewInvalidError("Fail"))
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Error(t, result.Error())
}

func TestCompletableCancel(t *testing.T) {
	completable := common.NewCompletable()

	err := completable.Cancel()
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Error(t, result.Error())
}

