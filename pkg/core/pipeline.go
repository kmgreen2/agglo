package core

type PipelineProcess interface {
	Process(in map[string]interface{}) (map[string]interface{}, error)
}
