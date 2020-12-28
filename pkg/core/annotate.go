package core

import "fmt"

// Annotation is a conditional annotation for a map
type Annotation struct {
	fieldKey string
	value string
	condition *Condition
}

// NewAnnotation will create a new conditional Annotation object based on the provided
// (key, value) and condition.  When used, if the condition is true, (key, value) is
// added to the provided map in the Process function
func NewAnnotation(key string, value string, condition *Condition) *Annotation {
	return &Annotation{
		key,
		value,
		condition,
	}
}

// ShouldAnnotate will return true if the underlying condition is satisfied in the
// provided map; otherwise, return false
func (a Annotation) ShouldAnnotate(in map[string]interface{}) (bool, error) {
	return a.condition.Evaluate(in)
}

// Annotate will annotate the provided map with the underlying annotation
func (a Annotation) Annotate(in map[string]interface{}) error {
	if _, ok := in[a.fieldKey]; ok {
		return fmt.Errorf("field exists: cannot annotate with field '%s'", a.fieldKey)
	}
	in[a.fieldKey] = a.value
	return nil
}

// Annotate is a pipeline processor that will conditionally apply underlying annotations to a provided map
type Annotate struct {
	annotations []*Annotation
}

// Process will conditionally apply underlying annotations to a copy of a provided map
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

// AnnotateBuilder is a builder for an Annotate pipeline processor
type AnnotateBuilder struct {
	annotate *Annotate
}

// NewAnnotateBuilder creates a new AnnotateBuilder
func NewAnnotateBuilder() *AnnotateBuilder {
	return &AnnotateBuilder{
		annotate: &Annotate{},
	}
}

// Add will add a new Annotation to the underlying AnnotateBuilder
func (builder *AnnotateBuilder) Add(annotation *Annotation) *AnnotateBuilder {
	builder.annotate.annotations = append(builder.annotate.annotations, annotation)
	return builder
}

// Build will build a new Annotate object from the underlying builder
func (builder *AnnotateBuilder) Build() *Annotate {
	return builder.annotate
}