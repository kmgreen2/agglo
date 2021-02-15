package process

import (
	"context"
	"regexp"
	"strings"
)

type Filter struct {
	name string
	regex       *regexp.Regexp
	keepMatched bool
}

func NewRegexKeyFilter(name string, expr string, keepMatched bool) (*Filter, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Filter {
		name: name,
		regex: re,
		keepMatched: keepMatched,
	}, nil
}

func NewListKeyFilter(name string, keys []string, keepMatched bool) (*Filter, error) {
	expr := strings.Join(keys, "|")
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return &Filter {
		name: name,
		regex: re,
		keepMatched: keepMatched,
	}, nil
}

func (filter Filter) Name() string {
	return filter.name
}

func (filter *Filter) process(in map[string]interface{}, out map[string]interface{}) interface{} {
	for k, v := range in {
		if filter.regex.Match([]byte(k)) == filter.keepMatched {
			switch vVal := v.(type) {
			case map[string]interface{}:
				out[k] = filter.process(vVal, make(map[string]interface{}))
			default:
				out[k] = v
			}
		}
	}
	return out
}

func (filter *Filter) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	out := make(map[string]interface{})

	filter.process(in, out)

	return out, nil
}


