package core

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"reflect"
	"regexp"
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
		outSlice := CopyableSlice(slice).DeepCopy()
		return &Transformable{outSlice}
	} else if t.Kind() == reflect.Map {
		m := t.Value().(map[string]interface{})
		outMap := CopyableMap(m).DeepCopy()
		return &Transformable{outMap}
	} else {
		return &Transformable{t.Value()}
	}
}

type FieldTransformation interface {
	Transform(in *Transformable) (*Transformable, error)
}

type MapTransformation struct {
	MapFunc func(interface{}) (interface{}, error)
}

func NewExecMapTransformation(path string) *MapTransformation {
	execRunnable := common.NewExecRunnable(path)

	return &MapTransformation{
		func (in interface{}) (interface{}, error) {
			err := execRunnable.SetArgs(in)
			if err != nil {
				return nil, err
			}
			return execRunnable.Run()
		},
	}
}

func (t MapTransformation) Transform(in *Transformable) (*Transformable, error) {
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

type LeftFoldTransformation struct {
	FoldFunc func(acc, v interface{}) (interface{}, error)
}

func NewExecLeftFoldTransformation(path string) *LeftFoldTransformation {
	execRunnable := common.NewExecRunnable(path)

	return &LeftFoldTransformation{
		func (acc, in interface{}) (interface{}, error) {
			switch val := in.(type) {
			case map[string]interface{}:
				if _, ok := val["agglo:acc"]; !ok {
					val["agglo:acc"] = acc
				} else {
					msg := fmt.Sprintf("reserved key 'agglo:acc' should not be set in args to fold")
					return nil, common.NewInvalidError(msg)
				}
			default:
				msg := fmt.Sprintf("expected map[string]interface{} argument to fold.  Got %v",
					reflect.TypeOf(val))
				return nil, common.NewInvalidError(msg)
			}

			err := execRunnable.SetArgs(in)
			if err != nil {
				return nil, err
			}
			return execRunnable.Run()
		},
	}
}

func (t LeftFoldTransformation) Transform(in *Transformable) (*Transformable, error) {
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

type RightFoldTransformation struct {
	FoldFunc func(acc, v interface{}) (interface{}, error)
}

func NewExecRightFoldTransformation(path string) *RightFoldTransformation {
	execRunnable := common.NewExecRunnable(path)

	return &RightFoldTransformation{
		func (acc, in interface{}) (interface{}, error) {
			switch val := in.(type) {
			case map[string]interface{}:
				if _, ok := val["agglo:acc"]; !ok {
					val["agglo:acc"] = acc
				} else {
					msg := fmt.Sprintf("reserved key 'agglo:acc' should not be set in args to fold")
					return nil, common.NewInvalidError(msg)
				}
			default:
				msg := fmt.Sprintf("expected map[string]interface{} argument to fold.  Got %v",
					reflect.TypeOf(val))
				return nil, common.NewInvalidError(msg)
			}

			err := execRunnable.SetArgs(in)
			if err != nil {
				return nil, err
			}
			return execRunnable.Run()
		},
	}
}

func (t RightFoldTransformation) Transform(in *Transformable) (*Transformable, error) {
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

type CopyTransformation struct {}
func (t CopyTransformation) Transform(in *Transformable) (*Transformable, error) {
	return in.Copy(), nil
}

type SumTransformation struct {}
func (t SumTransformation) Transform(in *Transformable) (*Transformable, error) {
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

/*
 * Fold helpers
 */
func foldMinFunc(acc, v interface{}) (interface{}, error) {
	if acc == nil {
		acc = v
		return acc, nil
	}
	accVal, vVal, err := NumericResolver(acc, v)
	if err != nil {
		return 0, err
	} else if accVal > vVal {
		accVal = vVal
	}
	return accVal, nil
}

func foldMaxFunc(acc, v interface{}) (interface{}, error) {
	if acc == nil {
		acc = v
		return acc, nil
	}
	accVal, vVal, err := NumericResolver(acc, v)
	if err != nil {
		return 0, err
	} else if accVal < vVal {
		accVal = vVal
	}
	return accVal, nil
}

func foldCountFunc(matcher func(interface{}) bool) func (acc, v interface{}) (interface{}, error) {
	return func(acc, v interface{}) (interface{}, error) {
		if acc == nil {
			acc = 0
			if matcher(v) {
				acc = float64(1)
			}
			return acc, nil
		}
		if accVal, err := GetNumeric(acc); err != nil {
			return 0, err
		} else {
			if matcher(v) {
				accVal++
			}
			return accVal, nil
		}
	}
}

// The fold functions are mostly for illustration.
// ToDo: Need to figure out the best way to serialize
// and generalize the matcher functions, so more useful
// folds can be done
var LeftFoldMin = &LeftFoldTransformation{
	foldMinFunc,
}
var RightFoldMin = &RightFoldTransformation{
	foldMinFunc,
}
var LeftFoldMax = &LeftFoldTransformation{
	foldMaxFunc,
}
var RightFoldMax = &RightFoldTransformation{
	foldMaxFunc,
}
var LeftFoldCountAll = &LeftFoldTransformation{
	foldCountFunc(func (x interface{}) bool {
		return true
	}),
}
var RightFoldCountAll = &RightFoldTransformation{
	foldCountFunc(func (x interface{}) bool {
		return true
	}),
}

/*
 * Map helpers
 */
func mapApplyRegex(regex string, replace string) func(interface{}) (interface{}, error) {
	// Not ideal, but we want to compile the regex and include in the returned function
	// This means that an invalid regex will lead to every call in map to return an error
	re, err := regexp.Compile(regex)
	return func(v interface{}) (interface{}, error) {
		var source string
		if err != nil {
			return nil, err
		}
		switch val := v.(type) {
		case string:
			source = val
		default:
			msg := fmt.Sprintf("expected string for regex source, got %v", reflect.TypeOf(v))
			return nil, common.NewInvalidError(msg)
		}
		result := string(re.ReplaceAll([]byte(source), []byte(replace)))
		return result, nil
	}
}

func mapAddConstant(x interface{}) func(interface{}) (interface{}, error) {
	return func(v interface{}) (interface{}, error) {
		xVal, vVal, err := NumericResolver(x, v)
		if err != nil {
			return 0, err
		} else {
			return xVal + vVal, nil
		}
	}
}

func mapMultConstant(x interface{}) func(interface{}) (interface{}, error) {
	return func(v interface{}) (interface{}, error) {
		xVal, vVal, err := NumericResolver(x, v)
		if err != nil {
			return 0, err
		} else {
			return xVal * vVal, nil
		}
	}
}

func MapApplyRegex(regex string, replace string) FieldTransformation {
	return &MapTransformation{
		mapApplyRegex(regex, replace),
	}
}

func MapAddConstant(x float64) FieldTransformation {
	return &MapTransformation{
		mapAddConstant(x),
	}
}

func MapMultConstant(x float64) FieldTransformation {
	return &MapTransformation{
		mapMultConstant(x),
	}
}

type Transformation struct {
	transformers []FieldTransformation
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

func NewTransformation(transformers []FieldTransformation, condition *Condition) *Transformation {
	if condition == nil {
		condition = TrueCondition
	}
	return &Transformation{
		transformers,
		condition,
	}
}

type TransformationBuilder struct {
	transformation *Transformation
}

func NewTransformationBuilder() *TransformationBuilder {
	return &TransformationBuilder{
		&Transformation{
			condition: TrueCondition,
		},
	}
}

func (t *TransformationBuilder) AddFieldTransformation(transformation FieldTransformation) *TransformationBuilder {
	t.transformation.transformers = append(t.transformation.transformers, transformation)
	return t
}

func (t *TransformationBuilder) AddCondition(condition *Condition) *TransformationBuilder {
	t.transformation.condition = condition
	return t
}


func (t *TransformationBuilder) Get() *Transformation {
	return t.transformation
}


