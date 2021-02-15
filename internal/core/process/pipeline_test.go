package process_test

import (
	"fmt"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/pkg/state"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPipelineBasic(t *testing.T) {
	kvStore := kvs.NewMemKVStore()
	testMaps := test.PipelineTestMapsOne()
	partitionID := gUuid.New()
	name := "default"
	builder := process.NewPipelineBuilder()

	// Annotate the inbound map with the proper partitionID and name
	// partitionID is meant to separate different organizations or departments
	// name is meant to partition different classes of messages (e.g. CI/CD, messaging, etc.)
	annotatorBuilder := process.NewAnnotatorBuilder("foo")
	annotatorBuilder.Add(core.NewAnnotation(string(common.PartitionIDKey), partitionID.String(), core.TrueCondition))
	annotatorBuilder.Add(core.NewAnnotation(string(common.ResourceNameKey), name, core.TrueCondition))
	builder.Add(annotatorBuilder.Build())

	// Annotate the different types of maps based on the existence of
	// certain fields

	// Annotate version control maps
	conditionBuilder := core.NewExistsExpressionBuilder()
	conditionBuilder.Add("author", core.Exists)
	conditionBuilder.Add("hash", core.Exists)
	conditionBuilder.Add("githash", core.NotExists)
	condition, err := core.NewCondition(conditionBuilder.Get())
	assert.Nil(t, err)

	annotatorBuilder = process.NewAnnotatorBuilder("foo")
	annotatorBuilder.Add(core.NewAnnotation("version-control", "git-dev", condition))
	builder.Add(annotatorBuilder.Build())

	// Annotate deployment maps
	conditionBuilder = core.NewExistsExpressionBuilder()
	conditionBuilder.Add("author", core.NotExists)
	conditionBuilder.Add("hash", core.NotExists)
	conditionBuilder.Add("githash", core.Exists)
	condition, err = core.NewCondition(conditionBuilder.Get())
	assert.Nil(t, err)

	annotatorBuilder = process.NewAnnotatorBuilder("foo")
	annotatorBuilder.Add(core.NewAnnotation("deploy", "circleci-dev", condition))
	builder.Add(annotatorBuilder.Build())

	// Aggregate the annotated version control maps on author
	aggregation := core.NewAggregation(core.NewFieldAggregation("author",
		core.AggDiscreteHistogram, []string{}))
	condition, err = core.NewCondition(core.NewComparatorExpression(core.Variable("version-control"), "git-dev",
		core.Equal))
	builder.Add(process.NewAggregator("fooAgg", aggregation, condition, state.NewKvStateStore(kvStore), false, true))

	// Create a completion that joins on commit hash
	completion := core.NewCompletion([]string{"hash", "githash"}, -1)
	completer := process.NewCompleter("default", completion, kvStore)
	builder.Add(completer)

	// Create transformations that set fields to be stored and tee them out to the kvstore
	transformer := process.NewTransformer("fooTransformer", nil, ".", ".", false)
	condition, err = core.NewCondition(core.NewComparatorExpression(core.Variable("version-control"), "git-dev",
		core.Equal))
	transformationBuilder := core.NewTransformationBuilder()
	transformationBuilder.AddFieldTransformation(&core.CopyTransformation{})
	copyTransformation := transformationBuilder.Get()
	transformer.AddSpec("author", "gitAuthor", copyTransformation)
	transformer.AddSpec("hash", "gitHash", copyTransformation)

	builder.Add(process.NewKVTee("kvTee", kvStore, condition, transformer, nil))

	// Spawn a job for each completed completion that adds the git hash to a list
	var spawnOutput []string
	runnable := test.NewFuncRunnable(func (arg map[string]interface{}) {
		spawnOutput = append(spawnOutput, arg["githash"].(string))
	})

	condition, err = core.NewCondition(core.NewComparatorExpression(core.Variable(common.InternalKeyFromPrefix(
		common.CompletionStatusPrefix, "default")),
		"complete",	core.Equal))

	builder.Add(process.NewSpawner("foo", core.NewLocalJob(runnable), condition, -1, true))

	// Use a filter to strip the internally added fields
	filter, err := process.NewRegexKeyFilter("foo", "^agglo.*", false)
	builder.Add(filter)

	pipeline := builder.Get()

	for _, m := range testMaps {
		out, err := pipeline.RunSync(m)
		fmt.Println(out)
		assert.Nil(t, err)
	}

	fmt.Println(spawnOutput)

}