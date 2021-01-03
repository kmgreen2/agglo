package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"reflect"
)

type Runnable interface {
	Run() (interface{}, error)
}

type PartialRunnable interface {
	Runnable
	SetArgs(args ...interface{}) error
}

// ExecRunnable will run a command that accepts a JSON-encoded map on stdin
// and returns a JSON-encoded map on stdout
type ExecRunnable struct {
	path string
	args map[string]interface{}
}

func NewExecRunnable(path string) *ExecRunnable {
	return &ExecRunnable{
		path: path,
	}
}

func NewExecRunnableWithArgs(path string, args map[string]interface{}) *ExecRunnable {
	return &ExecRunnable{
		path: path,
		args: args,
	}
}

func (runnable *ExecRunnable) Run() (interface{}, error) {
	var outMap map[string]interface{}
	cmd := exec.Command(runnable.path)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	encodeBuffer := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(encodeBuffer)
	err = encoder.Encode(runnable.args)
	if err != nil {
		return nil, err
	}

	go func() {
		_, _ = io.WriteString(stdin, encodeBuffer.String())
		_ = stdin.Close()
	}()

	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	decodeBuffer := bytes.NewBuffer(out)
	decoder := json.NewDecoder(decodeBuffer)
	err = decoder.Decode(&outMap)
	if err != nil {
		return nil, err
	}
	return outMap, err
}

func (runnable *ExecRunnable) SetArgs(args ...interface{}) error {
	if len(args) != 1 {
		msg := fmt.Sprintf("ExecRunnable expects a single map[string]interface{} argument.  Got %d args",
			len(args))
		return NewInvalidError(msg)
	}

	switch val := args[0].(type) {
	case map[string]interface{}:
		runnable.args = val
		return nil
	default:
		msg := fmt.Sprintf("ExecRunnable expects a single map[string]interface{} argument.  Got %v",
			reflect.TypeOf(val))
		return NewInvalidError(msg)
	}
}


