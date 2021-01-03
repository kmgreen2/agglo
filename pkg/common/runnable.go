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
	cmdArgs []string
	args interface{}
}

func NewExecRunnable(path string) *ExecRunnable {
	return &ExecRunnable{
		path: path,
	}
}

func NewExecRunnableWithCmdArgs(path string, cmdArgs ...string) *ExecRunnable {
	return &ExecRunnable{
		path: path,
		cmdArgs: cmdArgs,
	}
}

func NewExecRunnableWithArgs(path string, args interface{}) *ExecRunnable {
	return &ExecRunnable{
		path: path,
		args: args,
	}
}

func (runnable *ExecRunnable) Run() (interface{}, error) {
	var outMap map[string]interface{}
	cmd := exec.Command(runnable.path, runnable.cmdArgs...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	encodeBuffer := bytes.NewBuffer([]byte{})
	switch val := runnable.args.(type) {
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
	switch runnable.args.(type) {
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
	case []interface{}:
		runnable.args = val
		return nil
	case string:
		runnable.args = val
		return nil
	default:
		msg := fmt.Sprintf("ExecRunnable expects a single map[string]interface{} argument.  Got %v",
			reflect.TypeOf(val))
		return NewInvalidError(msg)
	}
}


