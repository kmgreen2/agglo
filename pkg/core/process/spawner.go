package process

import (
	"github.com/kmgreen2/agglo/pkg/core"
	"time"
)

type Spawner struct {
	job       core.Job
	condition *core.Condition
	delay     time.Duration
	doSync    bool
}

func NewSpawner(job core.Job, condition *core.Condition, delay time.Duration, doSync bool) *Spawner {
	return &Spawner{
		job: job,
		condition: condition,
		delay: delay,
		doSync: doSync,
	}
}

func (s Spawner) Process(in map[string]interface{}) (map[string]interface{}, error) {
	out := core.CopyableMap(in).DeepCopy()
	shouldRun, err := s.condition.Evaluate(out)
	if err != nil {
		return out, err
	}
	if shouldRun {
		f := s.job.Run(s.delay, s.doSync, out)
		if s.doSync {
			_, err := f.Get()
			if err != nil {
				return in, err
			}
		}
	}
	return out, nil
}
