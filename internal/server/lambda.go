package server

import (
	"fmt"
	"github.com/kmgreen2/agglo/internal/core/process"
)

type Request map[string]interface{}
type Response map[string]interface{}

func LambdaHandler(configBytes []byte) func (request Request) (Response, error) {
	return func (request Request) (Response, error) {
		var in, out map[string]interface{}
		in = request

		pipelines, err := process.PipelinesFromJson(configBytes)
		if err != nil {
			panic(fmt.Sprintf("%+v", err))
		}

		for _, pipeline := range pipelines.Underlying() {
			out, err = pipeline.RunSync(in)
			if err != nil {
				panic(err)
			}
		}
		return out, pipelines.Shutdown()
	}
}