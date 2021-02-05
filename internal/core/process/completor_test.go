package process_test

import (
	"context"
	"errors"
	gUuid "github.com/google/uuid"
	"github.com/kmgreen2/agglo/internal/common"
	"github.com/kmgreen2/agglo/pkg/util"
	"github.com/kmgreen2/agglo/internal/core"
	"github.com/kmgreen2/agglo/internal/core/process"
	"github.com/kmgreen2/agglo/pkg/kvs"
	"github.com/kmgreen2/agglo/test"
	"github.com/stretchr/testify/assert"
	gorand "math/rand"
	"strings"
	"testing"
	"time"
)


/**
 * NOTE: All of the completer tests run using doBasicCompleter intentionally test the completer
 * functions by serially calling Process() and Checkpoint().  This allows us to test the basic functionality.
 *
 * ToDo(KMG): Add doConcurrentCompleter, which will process maps in parallel, then call Checkpoint() and check the
 * final result.
 */
func doBasicCompleter(numMaps, numJoined int, timeout time.Duration, missingJoinKey string, forceTimeout bool) (int, int,
	int, error) {
	partitionID, err := gUuid.NewUUID()
	if err != nil {
		return 0, 0, 0, nil
	}

	maps, joinedKeys := test.GetJoinedMaps(numMaps, numJoined, partitionID, "foo")

	if len(missingJoinKey) > 0 {
		joinedKeys = append(joinedKeys, missingJoinKey)
	}

	completion := core.NewCompletion(joinedKeys, timeout)

	kvStore := kvs.NewMemKVStore()

	completer := process.NewCompleter("foo", completion, kvStore)

	numTriggered := 0
	numComplete := 0
	numTimedOut := 0
	for _, m := range maps {
		out, err := completer.Process(context.Background(), m)
		if err != nil {
			return 0, 0, 0, nil
		}
		if val, ok := out[common.InternalKeyFromPrefix(common.CompletionStatusPrefix, "foo")]; ok {
			// Timeout clock starts after first keepMatched for a join set
			if forceTimeout {
				time.Sleep(timeout)
				forceTimeout = false
			}

			if strings.Compare(val.(string), "complete") == 0 {
				numComplete++
			}
			if strings.Compare(val.(string), "triggered") == 0 {
				numTriggered++
			}
			if strings.Compare(val.(string), "timedout") == 0 {
				numTimedOut++
			}
		}
		_, _, err = completion.Match(m)
		if err != nil && errors.Is(err, &util.NotFoundError{}){
			continue
		} else if err != nil {
			return 0, 0, 0, err
		}
	}
	return numComplete, numTriggered, numTimedOut, nil
}

func TestCompleterHappyPath(t *testing.T) {
	numRuns := 64
	maxMaps := 128

	for i := 0; i < numRuns; i++ {
		numMaps := (gorand.Int() % (maxMaps - 2)) + 2
		numJoins := (gorand.Int() % numMaps) + 1

		numComplete, numTriggered, numTimedOut, err := doBasicCompleter(numMaps, numJoins, -1, "", false)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 1, numComplete)
		assert.Equal(t, numJoins-1, numTriggered)
		assert.Equal(t, 0, numTimedOut)
	}
}

func TestCompleterTimeoutNotify(t *testing.T) {
	numRuns := 4
	maxMaps := 128

	for i := 0; i < numRuns; i++ {
		// Need at least two maps and joins to timeout.  With one, it will just complete
		numMaps := (gorand.Int() % (maxMaps - 2)) + 2
		numJoins := numMaps

		numComplete, numTriggered, numTimedOut, err := doBasicCompleter(numMaps, numJoins,
			100*time.Millisecond, "", true)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 0, numComplete)
		assert.Equal(t, 1, numTriggered)
		assert.Equal(t, numMaps-1, numTimedOut)
	}
}

func TestCompleterIncomplete(t *testing.T) {
	numRuns := 64
	maxMaps := 128

	for i := 0; i < numRuns; i++ {
		// Need at least two maps and joins to timeout.  With one, it will just complete
		numMaps := (gorand.Int() % (maxMaps - 2)) + 2
		numJoins := numMaps

		numComplete, numTriggered, numTimedOut, err := doBasicCompleter(numMaps, numJoins,
			-1, gUuid.New().String(), false)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Equal(t, 0, numComplete)
		assert.Equal(t, numJoins, numTriggered)
		assert.Equal(t, 0, numTimedOut)
	}
}

