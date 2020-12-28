package core

import "time"

type TriggerParams map[string]interface{}

type Trigger interface {
	Pull(params TriggerParams, delay time.Duration, sync bool) error
}

type LocalTrigger struct {
	triggerFunc func(params TriggerParams) error
}

func NewLocalTrigger(triggerFunc func(params TriggerParams) error) *LocalTrigger {
	return &LocalTrigger{
		triggerFunc: triggerFunc,
	}
}

func (trigger LocalTrigger) Pull(params TriggerParams, delay time.Duration, sync bool) error {
	return nil
}