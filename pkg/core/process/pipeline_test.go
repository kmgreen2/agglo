package process_test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/stretchr/testify/assert"
	"testing"
)

// Test maps simulating commits to a code repo and deployments
func pipelineTestMapsOne() []map[string]interface{} {
	return []map[string]interface{} {
		{
			"author": "kmgreen2",
			"parent": "null",
			"hash": "abcd",
			"body": "first commit",
		},
		{
			"author": "kmgreen2",
			"parent": "abcd",
			"hash": "deff",
			"body": "second commit",
		},
		{
			"author": "foobar",
			"parent": "deff",
			"hash": "beef",
			"body": "third commit",
		},
		{
			"author": "foobar",
			"parent": "beef",
			"hash": "f00d",
			"body": "fourth commit",
		},
		{
			"user": "deploybot",
			"githash": "abcd",
		},
		{
			"user": "deploybot",
			"githash": "beef",
		},
	}
}

func TestPipelineBasic(t *testing.T) {
	testMaps := pipelineTestMapsOne()
	//partitionID := gUuid.New()
	//name := "test"
	builder := process.NewPipelineBuilder()

	// Annotate the different types of maps based on the existence of
	// certain fields

	// Annotate version control maps
	conditionBuilder := core.NewExistsExpressionBuilder()
	conditionBuilder.Add("author", core.Exists)
	conditionBuilder.Add("hash", core.Exists)
	conditionBuilder.Add("githash", core.NotExists)
	condition, err := core.NewCondition(conditionBuilder.Get())
	assert.Nil(t, err)

	annotatorBuilder := process.NewAnnotatorBuilder()
	annotatorBuilder.Add(core.NewAnnotation("version-control", "git-dev", condition))
	builder.Add(annotatorBuilder.Build())

	// Annotate deployment maps
	conditionBuilder = core.NewExistsExpressionBuilder()
	conditionBuilder.Add("author", core.NotExists)
	conditionBuilder.Add("hash", core.NotExists)
	conditionBuilder.Add("githash", core.Exists)
	condition, err = core.NewCondition(conditionBuilder.Get())
	assert.Nil(t, err)

	annotatorBuilder = process.NewAnnotatorBuilder()
	annotatorBuilder.Add(core.NewAnnotation("deploy", "circleci-dev", condition))
	builder.Add(annotatorBuilder.Build())


	pipeline := builder.Get()

	for _, m := range testMaps {
		out, err := pipeline.RunSync(m)
		fmt.Println(out)
		assert.Nil(t, err)
	}


}