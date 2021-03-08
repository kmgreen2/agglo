package process

import (
	"context"
	"fmt"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/pkg/util"
	"reflect"
	"time"
)

var SpawnMetadataKey string = string(common.SpawnMetadataKey)

type Spawner struct {
	name string
	job       core.Job
	condition *core.Condition
	delay     time.Duration
	doSync    bool
}

func NewSpawner(name string, job core.Job, condition *core.Condition, delay time.Duration, doSync bool) *Spawner {
	return &Spawner{
		name: name,
		job: job,
		condition: condition,
		delay: delay,
		doSync: doSync,
	}
}

func (s Spawner) Name() string {
	return s.name
}

func (s Spawner) Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error) {
	out := util.CopyableMap(in).DeepCopy()
	shouldRun, err := s.condition.Evaluate(out)
	if err != nil {
		return out, PipelineProcessError(s, err, "evaluating condition")
	}
	if shouldRun {
		f := s.job.Run(s.delay, s.doSync, out)
		if s.doSync {
			result := f.Get()
			if result.Error() != nil {
				return in, PipelineProcessError(s, result.Error(), "running job")
			}
			if _, ok := out[SpawnMetadataKey]; !ok {
				out[SpawnMetadataKey] = make([]map[string]interface{}, 0)
			}

			switch outVal := out[SpawnMetadataKey].(type) {
			case []map[string]interface{}:
				spawnResult := map[string]interface{} {
					s.name: result.Value(),
				}
				out[SpawnMetadataKey] = append(outVal, spawnResult)
			default:
				msg := fmt.Sprintf("detected corrupted %s in map when spawning.  expected []map[string]string, got %v",
					SpawnMetadataKey, reflect.TypeOf(outVal))
				return nil, util.NewInternalError(msg)
			}
		}
	}
	return out, nil
}
