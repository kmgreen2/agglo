package common

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"reflect"
)

type Runnable interface {
	Run(ctx context.Context) (interface{}, error)
}

type PartialRunnable interface {
	Runnable
	SetInData(inData interface{}) error
}

type ExecOption func(runnable *ExecRunnable)

func WithContext(ctx context.Context) ExecOption {
	return func (runnable *ExecRunnable) {
		runnable.ctx = ctx
	}
}
func WithCmdArgs(cmdArgs...string) ExecOption {
	return func (runnable *ExecRunnable) {
		runnable.cmdArgs = cmdArgs
	}
}
func WithInData(inData interface{}) ExecOption {
	return func (runnable *ExecRunnable) {
		runnable.inData = inData
	}
}
func WithPath(path string) ExecOption {
	return func (runnable *ExecRunnable) {
		runnable.path = path
	}
}

// ExecRunnable will run a command that accepts a JSON-encoded map on stdin
// and returns a JSON-encoded map on stdout
type ExecRunnable struct {
	ctx     context.Context
	path    string
	cmdArgs []string
	inData  interface{}
}

func NewExecRunnable(options ...ExecOption) *ExecRunnable {
	runnable := &ExecRunnable{
	}

	for _, option := range options {
		option(runnable)
	}

	return runnable
}

func (runnable *ExecRunnable) Run(ctx context.Context) (interface{}, error) {
	var outMap map[string]interface{}

	cmd := exec.Command(runnable.path, runnable.cmdArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	encodeBuffer := bytes.NewBuffer([]byte{})
	switch val := runnable.inData.(type) {
	case map[string]interface{}:
		encoder := json.NewEncoder(encodeBuffer)
		err = encoder.Encode(val)
		if err != nil {
			return nil, err
		}
	case []interface{}:
		encoder := json.NewEncoder(encodeBuffer)
		err = encoder.Encode(val)
		if err != nil {
			return nil, err
		}
	case string:
		encodeBuffer.Write([]byte(val))
	default:
		msg := fmt.Sprintf("ExecRunnable arg must be map[string]interface{}, []interface{} or string, got %v",
			reflect.TypeOf(val))
		return nil, NewInvalidError(msg)
	}

	go func() {
		_, _ = io.WriteString(stdin, encodeBuffer.String())
		_ = stdin.Close()
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return outMap, nil
	}

	decodeBuffer := bytes.NewBuffer(out)
	switch runnable.inData.(type) {
	case map[string]interface{}:
		decoder := json.NewDecoder(decodeBuffer)
		err = decoder.Decode(&outMap)
		if err != nil {
			return nil, err
		}
	case []interface{}:
		decoder := json.NewDecoder(decodeBuffer)
		err = decoder.Decode(&outMap)
		if err != nil {
			return nil, err
		}
	case string:
		decodeBuffer.Write([]byte(out))
	}

	return outMap, err
}

func (runnable *ExecRunnable) SetInData(inData interface{}) error {
	switch val := inData.(type) {
	case map[string]interface{}:
		runnable.inData = val
		return nil
	case []interface{}:
		runnable.inData = val
		return nil
	case string:
		runnable.inData = val
		return nil
	default:
		msg := fmt.Sprintf("ExecRunnable expects a single map[string]interface{} argument.  Got %v",
			reflect.TypeOf(val))
		return NewInvalidError(msg)
	}
}


