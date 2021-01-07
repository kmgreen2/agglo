package test

import (
	"fmt"
	"github.com/kmgreen2/agglo/pkg/common"
	"time"
	"reflect"
)

type SquareRunnable struct {
	value int
}

func NewSquareRunnable(value int) *SquareRunnable {
	return &SquareRunnable{
		value: value,
	}
}

func (r SquareRunnable) Run() (interface{}, error) {
	return r.value*r.value, nil
}

func (r *SquareRunnable) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("SquareRunnable should have 1 int arg, found %d args", len(args))
		return common.NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case int:
		r.value = rv
	default:
		msg := fmt.Sprintf("SquareRunnable should have an int arg, found %v", reflect.TypeOf(args[0]))
		return common.NewInvalidError(msg)
	}
	return nil
}

type SleepRunnable struct {
	value int
}

func NewSleepRunnable(value int) *SleepRunnable {
	return &SleepRunnable{
		value: value,
	}
}

func (r SleepRunnable) Run() (interface{}, error) {
	time.Sleep(time.Duration(r.value) * time.Second)
	return r.value, nil
}

func (r *SleepRunnable) SetArgs(args ...interface{}) error {
	return nil
}

type SleepAndFailRunnable struct {
	value int
}

func NewSleepAndFailRunnable(value int) *SleepAndFailRunnable {
	return &SleepAndFailRunnable{
		value: value,
	}
}

func (r SleepAndFailRunnable) Run() (interface{}, error) {
	if r.value > 0 {
		time.Sleep(time.Duration(r.value) * time.Second)
	}
	return nil, common.NewInvalidError("Failed")
}

func (r *SleepAndFailRunnable) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("SleepAndFailRunnable should have 1 int arg, found %d args", len(args))
		return common.NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case int:
		r.value = rv
	default:
		msg := fmt.Sprintf("SleepAndFailRunnable should have an int arg, found %v", reflect.TypeOf(args[0]))
		return common.NewInvalidError(msg)
	}
	return nil
}

type FailRunnable struct {

}

func NewFailRunnable() *FailRunnable {
	return &FailRunnable{}
}

func (r FailRunnable) Run() (interface{}, error) {
	return nil, common.NewInvalidError("Failed")
}

func (r *FailRunnable) SetArgs(args ...interface{}) error {
	return nil
}

type FailThenSucceedRunnable struct {
	numFails int
	currCalls int
}

func NewFailThenSucceedRunnable(numFails int) *FailThenSucceedRunnable {
	return &FailThenSucceedRunnable{
		numFails: numFails,
	}
}

func (r *FailThenSucceedRunnable) Run() (interface{}, error) {
	r.currCalls++
	if r.currCalls <= r.numFails {
		return nil, common.NewInvalidError("Failed")
	}
	return r.numFails, nil
}

func (r *FailThenSucceedRunnable) SetArgs(args ...interface{}) error {
	return nil
}

type FuncRunnable struct {
	runFunc func(arg map[string]interface{})
	arg map[string]interface{}
}

func (r FuncRunnable) Run() (interface{}, error) {
	r.runFunc(r.arg)
	return nil, nil
}

func (r *FuncRunnable) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("FuncRunnable should have 1 map[string]interface{} arg, found %d args", len(args))
		return common.NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case map[string]interface{}:
		r.arg = rv
	default:
		msg := fmt.Sprintf("FuncRunnable should have an map[string]interface{} arg, found %v", reflect.TypeOf(args[0]))
		return common.NewInvalidError(msg)
	}
	return nil
}

func NewFuncRunnable(runFunc func(arg map[string]interface{})) *FuncRunnable {
	return &FuncRunnable{
		runFunc: runFunc,
	}
}

