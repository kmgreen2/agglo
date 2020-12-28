package common

import (
	"fmt"
	"reflect"
	"time"
)

type PartialRunnable interface {
	Run() (interface{}, error)
	SetArgs(args ...interface{}) error
}

type Runnable interface {
	Run() (interface{}, error)
}

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
		return NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case int:
		r.value = rv
	default:
		msg := fmt.Sprintf("SquareRunnable should have an int arg, found %v", reflect.TypeOf(args[0]))
		return NewInvalidError(msg)
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
	return nil, NewInvalidError("Failed")
}

func (r *SleepAndFailRunnable) SetArgs(args ...interface{}) error {
	if len(args) > 1 {
		msg := fmt.Sprintf("SleepAndFailRunnable should have 1 int arg, found %d args", len(args))
		return NewInvalidError(msg)
	}
	switch rv := args[0].(type) {
	case int:
		r.value = rv
	default:
		msg := fmt.Sprintf("SleepAndFailRunnable should have an int arg, found %v", reflect.TypeOf(args[0]))
		return NewInvalidError(msg)
	}
	return nil
}

type FailRunnable struct {

}

func NewFailRunnable() *FailRunnable {
	return &FailRunnable{}
}

func (r FailRunnable) Run() (interface{}, error) {
	return nil, NewInvalidError("Failed")
}

func (r *FailRunnable) SetArgs(args ...interface{}) error {
	return nil
}



