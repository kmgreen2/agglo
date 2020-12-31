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