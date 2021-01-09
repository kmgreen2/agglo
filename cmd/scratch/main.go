package main

import (
	"context"
	"github.com/kmgreen2/agglo/pkg/observability"
)

func main() {
	exporter, err := observability.NewExporter(observability.StdoutExporter)
	if err != nil {
		panic(err)
	}
	_ = exporter.Start()
	defer func() {_ = exporter.Stop()}()

	emitter := observability.NewEmitter("foo.io")
	emitter.AddMetric("fizz", observability.Int64Counter)
	emitter.AddMetric("buzz", observability.Int64Recorder)
	ctx1, span1 := emitter.CreateSpan(context.Background(), "foo")
	ctx2, span2 := emitter.CreateSpan(ctx1, "bar")
	_, span3 := emitter.CreateSpan(ctx2, "baz")

	span1.AddEvent("Did a thing")

	emitter.AddInt64("fizz", 1)
	emitter.AddInt64("fizz", 1)
	emitter.AddInt64("fizz", 1)
	emitter.AddInt64("fizz", 1)
	emitter.AddInt64("fizz", 1)
	emitter.AddInt64("fizz", 1)

	span1.End()
	span2.End()
	span3.End()
}

