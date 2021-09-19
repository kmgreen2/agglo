package search_test

import (
	"github.com/kmgreen2/agglo/pkg/search"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBasicElasticQueryCompleteHappyPath(t *testing.T) {
	startTime:= time.Date(2020, 10, 1, 4,0,0, 0, time.UTC)
	endTime:= time.Date(2020, 10, 2, 4,0,0, 0, time.UTC)
	builder := search.NewSimpleElasticQueryBuilder()
	expected := "{\"query\":{\"bool\":{\"filter\":{\"range\":{\"created\":{\"gte\":\"2020-10-01 04:00:00\",\"lte\":\"2020-10-02 04:00:00\"}}},\"must\":{\"nested\":{\"path\":\"keywords\",\"query\":[{\"bool\":{\"minimum_should_match\":2,\"should\":[{\"term\":{\"keywords.id\":\"foo\"}},{\"term\":{\"keywords.value\":\"buzz\"}}]}},{\"bool\":{\"minimum_should_match\":2,\"should\":[{\"term\":{\"keywords.id\":\"fizz\"}},{\"term\":{\"keywords.value\":\"bar\"}}]}}]}},\"should\":{\"match\":{\"fulltext\":{\"query\":\"all of the things\"}}}}}}\n"

	elasticQuery := builder.
		AddTerm("foo", "buzz").
		AddTerm("fizz", "bar").AddText("all of the things").
		SetStartTime(startTime.Unix()).
		SetEndTime(endTime.Unix()).Build()

	jsonBytes, err := elasticQuery.Render()

	if err != nil {
		assert.Fail(t, err.Error())
	}

	assert.JSONEq(t, expected, string(jsonBytes))
}

