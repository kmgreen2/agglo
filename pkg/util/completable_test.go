package util_test

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompletableSuccess(t *testing.T) {
	completable := util.NewCompletable()

	err := completable.Success(context.Background(), int(1337))
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Nil(t, result.Error())
	assert.Equal(t, 1337, result.Value().(int))
}

func TestCompletableFail(t *testing.T) {
	completable := util.NewCompletable()

	err := completable.Fail(context.Background(), util.NewInvalidError("Fail"))
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Error(t, result.Error())
}

func TestCompletableCancel(t *testing.T) {
	completable := util.NewCompletable()

	err := completable.Cancel(context.Background())
	assert.Nil(t, err)

	result := completable.Future().Get()
	assert.Error(t, result.Error())
}

