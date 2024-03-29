package process

import (
	"context"
	"errors"
	"fmt"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"strings"
)

type Transformer struct {
	name string
	specs          []*TransformerSpec
	fieldSeparator string
	indexSeparator string
	forwardInputFields bool
}

type TransformerSpec struct {
	sourceField string
	targetField string
	transformation *core.Transformation
}

func NewTransformer(name string, specs []*TransformerSpec, fieldSeparator, indexSeparator string,
	forwardInputFields bool) *Transformer {
	return &Transformer{
		name,
		specs,
		fieldSeparator,
		indexSeparator,
		forwardInputFields,
	}
}

func (t Transformer) Name() string {
	return t.name
}

func DefaultTransformer() *Transformer {
	return NewTransformer("default", nil, ".", ".", false)
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
			if currValue, ok := curr[fieldName]; ok {
				curr = currValue.(map[string]interface{})
			} else {
				msg := fmt.Sprintf("could not find '%s' in the map", key)
				return nil, util.NewNotFoundError(msg)
			}
		} else if _, ok := curr[fieldName]; !ok {
			msg := fmt.Sprintf("could not find '%s' in the map", key)
			return nil, util.NewNotFoundError(msg)
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

	// Note(KMG): This path is specifically for copying entire input payloads
	if len(sourceField) == 0 && len(targetField) == 0 {
		result, err := transformation.Transform(core.NewTransformable(in))
		if err != nil {
			return err
		}

		switch rVal := result.Value().(type) {
		case map[string]interface{}:
			for k, v := range rVal {
				out[k] = v
			}
			return nil
		default:
			msg := fmt.Sprintf("source and target fields must be specified when transforming non-root maps")
			return util.NewInvalidError(msg)
		}
	}

	sourceDict, err := t.dictFromPath(sourceField, in)
	if err != nil {
		return err
	}


	fieldNames := strings.Split(targetField, t.fieldSeparator)
	var curr map[string]interface{} = out
	for i, fieldName := range fieldNames {
		should, err := transformation.ShouldTransform(util.Flatten(in))
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

func (t Transformer) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	var out map[string]interface{}
	if t.forwardInputFields {
		out = util.CopyableMap(in).DeepCopy()
	} else {
		out = make(map[string]interface{})
	}

	for _, spec := range t.specs {
		err := t.createPathAndTransform(spec.sourceField, spec.targetField, spec.transformation, in, out)
		// This will skip transforming fields when the source field cannot be found in the `in` map
		if err != nil && !errors.Is(err, &util.NotFoundError{}){
			return in, PipelineProcessError(t, err, "transforming fields")
		}
	}
	return out, nil
}
