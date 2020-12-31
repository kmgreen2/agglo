package pipeline

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"strings"
)

type Transformer struct {
	specs          []*TransformerSpec
	fieldSeparator string
	indexSeparator string
}

type TransformerSpec struct {
	sourceField string
	targetField string
	transformation *core.Transformation
}

func NewTransformer(specs []*TransformerSpec, fieldSeparator, indexSeparator string) *Transformer {
	return &Transformer{
		specs,
		fieldSeparator,
		indexSeparator,
	}
}

func (t *Transformer) AddSpec(sourceField, targetField string, transformation *core.Transformation) {
	t.specs = append(t.specs, &TransformerSpec{
		sourceField: sourceField,
		targetField: targetField,
		transformation: transformation,
	})
}

func (t Transformer) dictFromPath(key string, in map[string]interface{}) (map[string]interface{}, error) {
	fieldNames := strings.Split(key, t.fieldSeparator)
	var curr map[string]interface{} = in
	for i, fieldName := range fieldNames {
		if i != len(fieldNames) - 1 {
			curr = curr[fieldName].(map[string]interface{})
		}
	}
	return curr, nil
}

func (t Transformer) valueFromPath(key string, in map[string]interface{}) (interface{}, error) {
	fieldNames := strings.Split(key, t.fieldSeparator)
	var curr map[string]interface{} = in
	for i, fieldName := range fieldNames {
		if i != len(fieldNames) - 1 {
			curr = curr[fieldName].(map[string]interface{})
		}
	}
	return curr[fieldNames[len(fieldNames)-1]], nil
}

func (t Transformer) createPathAndTransform(sourceField, targetField string, transformation *core.Transformation, in,
	out map[string]interface{}) error {

	sourceDict, err := t.dictFromPath(sourceField, in)
	if err != nil {
		return err
	}
	fieldNames := strings.Split(targetField, t.fieldSeparator)
	var curr map[string]interface{} = out
	for i, fieldName := range fieldNames {
		should, err := transformation.ShouldTransform(core.Flatten(in))
		if err != nil {
			return err
		}
		if !should {
			continue
		}
		if _, ok := curr[fieldName]; !ok {
			curr[fieldName] = make(map[string]interface{})
		}
		if i < len(fieldNames) - 1 {
			curr = curr[fieldName].(map[string]interface{})
		} else {
			sourceFields := strings.Split(sourceField, t.fieldSeparator)
			sourceField := sourceFields[len(sourceFields)-1]
			result, err := transformation.Transform(core.NewTransformable(sourceDict[sourceField]))
			if err != nil {
				return err
			}
			curr[fieldName] = result.Value()
		}
	}

	return nil
}

func (t Transformer) Process(in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	for _, spec := range t.specs {
		err := t.createPathAndTransform(spec.sourceField, spec.targetField, spec.transformation, in, out)
		if err != nil {
			return in, err
		}
	}
	return out, nil
}
