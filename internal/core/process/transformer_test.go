package process_test

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyAllTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]", false)
	transformer.AddSpec("", "", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson, out)
}

func TestCopySomeTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]", false)
	transformer.AddSpec("a", "foo", transformation)
	transformer.AddSpec("b", "bar", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson["a"], out["foo"])
	assert.Equal(t, testJson["b"], out["bar"])
}

func TestCopyAllExplicitTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]", true)
	transformer.AddSpec("a", "foo", transformation)
	transformer.AddSpec("b", "bar", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson["a"], out["foo"])
	assert.Equal(t, testJson["b"], out["bar"])
	for k, v := range testJson {
		assert.Equal(t, v, out[k])
	}
}

func TestCopyAllAndTransformTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]", false)
	transformer.AddSpec("", "", transformation)
	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)

	for k, v := range testJson {
		assert.Equal(t, v, out[k])
	}
	assert.Equal(t, float64(12), out["bSum"])
}

func TestCopySomeAndTransformTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]", false)
	transformer.AddSpec("a", "foo", transformation)
	transformer.AddSpec("b", "bar", transformation)
	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson["a"], out["foo"])
	assert.Equal(t, testJson["b"], out["bar"])
	assert.Equal(t, float64(12), out["bSum"])
}

func TestCopyNoneAndTransformTransformer(t *testing.T) {
	testJson := test.TestJson()
	transformer := process.NewTransformer(nil, ".", "[]", false)
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.RightFoldMax).Get()
	transformer.AddSpec("b.d", "bMax", transformation)

	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.LeftFoldCountAll).Get()
	transformer.AddSpec("e", "eCount", transformation)

	out, err := transformer.Process(context.Background(), testJson)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(out))
	assert.Equal(t, float64(12), out["bSum"])
	assert.Equal(t, float64(5), out["bMax"])
	assert.Equal(t, float64(1), out["eCount"])
}
