package util_test

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestExecRunnable(t *testing.T) {
	testJson := test.TestJson()
	runnable := util.NewExecRunnable(util.WithPath("cat"))

	err := runnable.SetInData(testJson)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	result, err := runnable.Run(context.Background())
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	switch val := result.(type) {
	case map[string]interface{}:
		assert.Equal(t, testJson, val)
	default:
		assert.Fail(t, fmt.Sprintf("expected map[string]interface{}.  Got %v", reflect.TypeOf(val)))
	}
}

func TestExecRunnableFail(t *testing.T) {
	testJson := test.TestJson()
	runnable := util.NewExecRunnable(util.WithPath("foo/doesnotexist"))

	err := runnable.SetInData(testJson)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = runnable.Run(context.Background())
	assert.Error(t, err)
}

func TestExecRunnableUnexpectedOutput(t *testing.T) {
	testJson := test.TestJson()
	runnable := util.NewExecRunnable(util.WithPath("echo"))

	err := runnable.SetInData(testJson)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	_, err = runnable.Run(context.Background())
	assert.Error(t, err)
}