package core

import "time"

type Spawner struct {
	job Job
	condition *Condition
	delay time.Duration
	doSync bool
}

func NewSpawner(job Job, condition *Condition, delay time.Duration, doSync bool) *Spawner {
	return &Spawner{
		job: job,
		condition: condition,
		delay: delay,
		doSync: doSync,
	}
}

func (s Spawner) Process(in map[string]interface{}) (map[string]interface{}, error) {
	shouldRun, err := s.condition.Evaluate(in)
	if err != nil {
		return in, err
	}
	if shouldRun {
		f := s.job.Run(s.delay, s.doSync)
		if s.doSync {
			_, err := f.Get()
			if err != nil {
				return in, err
			}
		}
	}
	return in, nil
}
