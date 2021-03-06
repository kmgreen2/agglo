package process

import (
	"context"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/util"
)

// Annotator is a process processor that will conditionally apply underlying annotations to a provided map
type Annotator struct {
	name string
	annotations []*core.Annotation
}

// Process will conditionally apply underlying annotations to a copy of a provided map
func (a Annotator) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	out := util.CopyableMap(in).DeepCopy()
	for _, annotation := range a.annotations {
		should, err := annotation.ShouldAnnotate(util.Flatten(out))
		if err != nil {
			return in, PipelineProcessError(a, err, "evaluating condition")
		}
		if should {
			err = annotation.Annotate(out)
			if err != nil {
				return in, PipelineProcessError(a, err, "adding annotation")
			}
		}
	}

	return out, nil
}

func (a Annotator) Name() string {
	return a.name
}

// AnnotatorBuilder is a builder for an Annotator process processor
type AnnotatorBuilder struct {
	annotate *Annotator
}

// NewAnnotatorBuilder creates a new AnnotatorBuilder
func NewAnnotatorBuilder(name string) *AnnotatorBuilder {
	return &AnnotatorBuilder{
		annotate: &Annotator{
			name: name,
		},
	}
}

// Add will add a new Annotation to the underlying AnnotatorBuilder
func (builder *AnnotatorBuilder) Add(annotation *core.Annotation) *AnnotatorBuilder {
	builder.annotate.annotations = append(builder.annotate.annotations, annotation)
	return builder
}

// Build will build a new Annotator object from the underlying builder
func (builder *AnnotatorBuilder) Build() *Annotator {
	return builder.annotate
}