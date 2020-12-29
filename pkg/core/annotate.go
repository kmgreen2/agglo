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

// Annotator will annotate the provided map with the underlying annotation
func (a Annotation) Annotate(in map[string]interface{}) error {
	if _, ok := in[a.fieldKey]; ok {
		return fmt.Errorf("field exists: cannot annotate with field '%s'", a.fieldKey)
	}
	in[a.fieldKey] = a.value
	return nil
}

// Annotator is a pipeline processor that will conditionally apply underlying annotations to a provided map
type Annotator struct {
	annotations []*Annotation
}

// Process will conditionally apply underlying annotations to a copy of a provided map
func (a Annotator) Process(in map[string]interface{}) (map[string]interface{}, error) {
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

// AnnotatorBuilder is a builder for an Annotator pipeline processor
type AnnotatorBuilder struct {
	annotate *Annotator
}

// NewAnnotateBuilder creates a new AnnotatorBuilder
func NewAnnotateBuilder() *AnnotatorBuilder {
	return &AnnotatorBuilder{
		annotate: &Annotator{},
	}
}

// Add will add a new Annotation to the underlying AnnotatorBuilder
func (builder *AnnotatorBuilder) Add(annotation *Annotation) *AnnotatorBuilder {
	builder.annotate.annotations = append(builder.annotate.annotations, annotation)
	return builder
}

// Build will build a new Annotator object from the underlying builder
func (builder *AnnotatorBuilder) Build() *Annotator {
	return builder.annotate
}