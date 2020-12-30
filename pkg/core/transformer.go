package core

import (
	"fmt"
	"reflect"
	"strings"
)

type Transformable struct {
	value interface{}
}

func NewTransformable(value interface{}) *Transformable {
	return &Transformable{value}
}

func (t Transformable) Kind() reflect.Kind {
	return reflect.TypeOf(t.value).Kind()
}

func (t Transformable) Value() interface{} {
	return t.value
}

func (t Transformable) Copy() *Transformable {
	if t.Kind() == reflect.Slice {
		slice := t.Value().([]interface{})
		var outSlice []interface{}
		for _, v := range slice {
			outSlice = append(outSlice, v)
		}
		return &Transformable{outSlice}
	} else if t.Kind() == reflect.Map {
		m := t.Value().(map[string]interface{})
		outMap := make(map[string]interface{})
		for k, v := range m {
			outMap[k] = v
		}
		return &Transformable{outMap}
	} else {
		return &Transformable{t.Value()}
	}
}

type FieldTransformer interface {
	Transform(in *Transformable) (*Transformable, error)
}

type MapTransformer struct {
	MapFunc func(interface{}) (interface{}, error)
}
func (t MapTransformer) Transform(in *Transformable) (*Transformable, error) {
	if in.Kind() == reflect.Slice {
		slice := in.Value().([]interface{})
		var outSlice []interface{}
		for _, v := range slice {
			mVal, err := t.MapFunc(v)
			if err != nil {
				return nil, err
			}
			outSlice = append(outSlice, mVal)
		}
		return &Transformable{outSlice}, nil
	} else if in.Kind() == reflect.Map {
		m := in.Value().(map[string]interface{})
		outMap := make(map[string]interface{})
		for k, v := range m {
			mVal, err := t.MapFunc(v)
			if err != nil {
				return nil, err
			}
			outMap[k] = mVal
		}
		return &Transformable{outMap}, nil
	} else {
		mVal, err := t.MapFunc(in.Value())
		if err != nil {
			return nil, err
		}
		return &Transformable{mVal}, nil
	}
}

type LeftFoldTransformer struct {
	FoldFunc func(acc, v interface{}) (interface{}, error)
}
func (t LeftFoldTransformer) Transform(in *Transformable) (*Transformable, error) {
	if in.Kind() == reflect.Slice {
		slice := in.Value().([]interface{})
		var acc interface{}
		var err error
		for _, v := range slice {
			acc, err = t.FoldFunc(acc, v)
			if err != nil {
				return nil, err
			}
		}
		return &Transformable{acc}, nil
	} else {
		return nil, fmt.Errorf("")
	}
}

type RightFoldTransformer struct {
	FoldFunc func(acc, v interface{}) (interface{}, error)
}
func (t RightFoldTransformer) Transform(in *Transformable) (*Transformable, error) {
	if in.Kind() == reflect.Slice {
		slice := in.Value().([]interface{})
		var acc interface{}
		var err error
		for i, _ := range slice {
			acc, err = t.FoldFunc(acc, slice[len(slice)-i-1])
			if err != nil {
				return nil, err
			}
		}
		return &Transformable{acc}, nil
	} else {
		return nil, fmt.Errorf("")
	}
}

type CopyTransformer struct {}
func (t CopyTransformer) Transform(in *Transformable) (*Transformable, error) {
	return in.Copy(), nil
}

type SumTransformer struct {}
func (t SumTransformer) Transform(in *Transformable) (*Transformable, error) {
	if in.Kind() != reflect.Slice {
		return nil, fmt.Errorf("")
	}
	list := in.Value().([]interface{})
	sum := float64(0)
	for _, elm := range list {
		x, err := GetNumeric(elm)
		if err != nil {
			return nil, fmt.Errorf("")
		}
		sum += x
	}
	return &Transformable{sum}, nil
}

type Transformation struct {
	sourceField string
	transformers []FieldTransformer
	condition *Condition
}

func (t *Transformation) Transform(in *Transformable) (*Transformable, error) {
	var err error
	curr := in
	for _, transformer := range t.transformers {
		curr, err = transformer.Transform(curr)
		if err != nil {
			return nil, err
		}
	}
	return curr, nil
}

func (t *Transformation) ShouldTransform(in map[string]interface{}) (bool, error) {
	return t.condition.Evaluate(in)
}

func NewTransformation(source string, transformers []FieldTransformer, condition *Condition) *Transformation {
	if condition == nil {
		condition = TrueCondition
	}
	return &Transformation{
		source,
		transformers,
		condition,
	}
}

type Transformer struct {
	spec map[string]*Transformation
	fieldSeparator string
	indexSeparator string
}

func NewTransformer(spec map[string]*Transformation, fieldSeparator, indexSeparator string) *Transformer {
	if spec == nil {
		spec = make(map[string]*Transformation)
	}
	return &Transformer{
		spec,
		fieldSeparator,
		indexSeparator,
	}
}

func (t *Transformer) AddSpec(tgtField string, transformation *Transformation) {
	t.spec[tgtField] = transformation
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

func (t Transformer) createPathAndTransform(tgtKey string, transformation *Transformation, in,
	out map[string]interface{}) error {

	sourceDict, err := t.dictFromPath(transformation.sourceField, in)
	if err != nil {
		return err
	}
	fieldNames := strings.Split(tgtKey, t.fieldSeparator)
	var curr map[string]interface{} = out
	for i, fieldName := range fieldNames {
		should, err := transformation.ShouldTransform(Flatten(in))
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
			sourceFields := strings.Split(transformation.sourceField, t.fieldSeparator)
			sourceField := sourceFields[len(sourceFields)-1]
			result, err := transformation.Transform(&Transformable{sourceDict[sourceField]})
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

	for tgt, transformation := range t.spec {
		err := t.createPathAndTransform(tgt, transformation, in, out)
		if err != nil {
			return in, err
		}
	}
	return out, nil
}
