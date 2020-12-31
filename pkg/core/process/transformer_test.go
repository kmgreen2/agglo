package process_test

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"github.com/kmgreen2/agglo/pkg/core/process"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyAllTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]")
	transformer.AddSpec("", "", transformation)

	out, err := transformer.Process(testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson, out)
}

func TestCopySomeTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]")
	transformer.AddSpec("a", "foo", transformation)
	transformer.AddSpec("b", "bar", transformation)

	out, err := transformer.Process(testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson["a"], out["foo"])
	assert.Equal(t, testJson["b"], out["bar"])
}

func TestCopyAllAndTransformTransformer(t *testing.T) {
	testJson := test.TestJson()
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.CopyTransformation{}).Get()
	transformer := process.NewTransformer(nil, ".", "[]")
	transformer.AddSpec("", "", transformation)
	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	out, err := transformer.Process(testJson)
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
	transformer := process.NewTransformer(nil, ".", "[]")
	transformer.AddSpec("a", "foo", transformation)
	transformer.AddSpec("b", "bar", transformation)
	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	out, err := transformer.Process(testJson)
	assert.Nil(t, err)
	assert.Equal(t, testJson["a"], out["foo"])
	assert.Equal(t, testJson["b"], out["bar"])
	assert.Equal(t, float64(12), out["bSum"])
}

func TestCopyNoneAndTransformTransformer(t *testing.T) {
	testJson := test.TestJson()
	transformer := process.NewTransformer(nil, ".", "[]")
	builder := core.NewTransformationBuilder()
	transformation := builder.AddFieldTransformation(core.SumTransformation{}).Get()
	transformer.AddSpec("b.d", "bSum", transformation)

	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.RightFoldMax).Get()
	transformer.AddSpec("b.d", "bMax", transformation)

	builder = core.NewTransformationBuilder()
	transformation = builder.AddFieldTransformation(core.LeftFoldCountAll).Get()
	transformer.AddSpec("e", "eCount", transformation)

	out, err := transformer.Process(testJson)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(out))
	assert.Equal(t, float64(12), out["bSum"])
	assert.Equal(t, float64(5), out["bMax"])
	assert.Equal(t, float64(1), out["eCount"])
}
