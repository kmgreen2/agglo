package core

import "fmt"

type Annotation struct {
	fieldKey string
	value string
	condition *Condition
}

func NewAnnotation(key string, value string, condition *Condition) *Annotation {
	return &Annotation{
		key,
		value,
		condition,
	}
}

func (a Annotation) ShouldAnnotate(in map[string]interface{}) (bool, error) {
	return a.condition.Evaluate(in)
}

func (a Annotation) Annotate(in map[string]interface{}) error {
	if _, ok := in[a.fieldKey]; ok {
		return fmt.Errorf("field exists: cannot annotate with field '%s'", a.fieldKey)
	}
	in[a.fieldKey] = a.value
	return nil
}

type Annotate struct {
	annotations []*Annotation
}

func (a Annotate) Process(in map[string]interface{}) (map[string]interface{}, error) {
	out := CopyableMap(in).DeepCopy()
	for _, annotation := range a.annotations {
		should, err := annotation.ShouldAnnotate(Flatten(out))
		if err != nil {
			return nil, err
		}
		if should {
			err = annotation.Annotate(out)
			if err != nil {
				return nil, err
			}
		}
	}

	return out, nil
}

type AnnotateBuilder struct {
	annotate *Annotate
}

func NewAnnotateBuilder() *AnnotateBuilder {
	return &AnnotateBuilder{
		annotate: &Annotate{},
	}
}

func (builder *AnnotateBuilder) Add(annotation *Annotation) *AnnotateBuilder {
	builder.annotate.annotations = append(builder.annotate.annotations, annotation)
	return builder
}

func (builder *AnnotateBuilder) Build() *Annotate {
	return builder.annotate
}