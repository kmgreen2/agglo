package process_test

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewRegexKeyFilter(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"fi:zz": "bar",
		"baz": "foo",
	}

	filter, err := process.NewRegexKeyFilter(`^f.*`, true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	out, err := filter.Process(context.Background(), testMap)
	assert.Equal(t, 2, len(out))
	assert.Equal(t, 1, out["foo"])
	assert.Equal(t, "bar", out["fi:zz"])
}

func TestNewRegexKeyFilterInvert(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"fi:zz": "bar",
		"baz": "foo",
	}

	filter, err := process.NewRegexKeyFilter(`^f.*`, false)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	out, err := filter.Process(context.Background(), testMap)
	assert.Equal(t, 1, len(out))
	assert.Equal(t, "foo", out["baz"])
}

func TestNewListKeyFilter(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"fizz": "bar",
		"baz": "foo",
	}

	filter, err := process.NewListKeyFilter([]string{"foo", "baz"}, true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	out, err := filter.Process(context.Background(), testMap)
	assert.Equal(t, 2, len(out))
	assert.Equal(t, 1, out["foo"])
	assert.Equal(t, "foo", out["baz"])
}

func TestNewListMultilevelKeyFilter(t *testing.T) {
	testMap := map[string]interface{} {
		"foo": 1,
		"fizz": "bar",
		"baz": "foo",
		"booze" : map[string]interface{}{
			"foo": 25,
			"bizz": "hi",
		},
		"blaze" : map[string]interface{}{
			"foo": "hi",
		},
	}

	filter, err := process.NewListKeyFilter([]string{"foo", "baz", "booze"}, true)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	out, err := filter.Process(context.Background(), testMap)
	assert.Equal(t, 3, len(out))
	assert.Equal(t, 1, out["foo"])
	assert.Equal(t, "foo", out["baz"])
	assert.Equal(t, 25, out["booze"].(map[string]interface{})["foo"])
}