# Pipeline Processes

All processes must implement the following interface:

```
type PipelineProcess interface {
	Process(ctx context.Context, in map[string]interface{}) (map[string]interface{}, error)
}
```

In general, a process will perform one or more actions based on the input map
`in`.  While the input map can technically be anything serialized to/from
`map[string]interface{}`, we currently support JSON.  Each process will also
output a map, which is the input to the next process in the pipeline, or gets
dropped on the floor if it is the last process in a pipeline.  The input to the
first process is the raw event posted to or read by binge.

There are currently 9 process types:

- **Annotation**: Conditionally add one or more annotations to the map

  Example: Add `{"foo": "bar"}` to any input map that contains key `bizz` with value `buzz`.

    Input:
  ```json
  {
      "bizz": "buzz",
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 1
      }
  }
  ```
 
  Output:
  ```json
  {
      "foo": "bar",
      "bizz": "buzz",
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 1
      }
  }
  ```

- **Aggregation**: Aggregate one or more fields (e.g. sum, max, histogram, etc.)

  _Aggregation Types_: Sum, Max, Min, Avg, Count, Histogram
 
  Example: Compute a sum of `someMap.someNum` and a histogram on `name`.  Note there are
  two successive input messages and the different outputs for each.  Note that the output
  contains the current value of the aggregation.  The location of the binge instances and
  the state store will dictate the consistency of the current values.
 
    Input 1:
    ```json
    {
        "name": "frank",
        "someMap": {
            "hi": "hello",
            "someNum": 1
        }
    }
    ```
   Input 2:
   ```json
   {
       "name": "jane",
       "someMap": {
           "hi": "hello",
           "someNum": 4
       }
   }
   ```
 
  Output 1:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 1
      },
      "internal:aggregation:PipelineName": {
          "someNum:AggSum": {
            "value": 1
          },
          "name:AggHistogram": {
              "frank": 1
            }
      }
  }
  ```
 
  Output 2:
  ```json
  {
      "name": "jane",
      "someMap": {
          "hi": "hello",
          "someNum": 4
      },
      "internal:aggregation:PipelineName": {
          "someNum:AggSum": {
            "value": 5
          },
          "name:AggHistogram": {
              "frank": 1,
              "jane": 1
            }
      }
  }
  ```
 
- **Completion**: Emit a completion event when the value of two or more specified fields is equal.  State is maintained
by a key-value store specified in the binge configuration.

  Example: Define a completion event on `head` and `cicdVcsHash` where `state` is `passed`.  Here we assume that
  both input 1 is processed before input 2.  Also, notice that the output of this process
  includes the current state of the completion.
 
  Input 1:
  ```json
  {
      "repo": "agglo",
      "head": "deadbeef",
      "state": "passed"
  }
   ```
  Input 2:
  ```json
  {
      "workflow": "agglo",
      "cicdVcsHash": "deadbeef",
      "state": "passed"
  }
  ```
  Output 1:
  ```json
  {
      "repo": "agglo",
      "head": "deadbeef",
      "state": "passed",
      "internal:completion:PipelineName": {
        "completionName": {
          "state": "triggered",
          "resolved": {
            "head": true
          },
          "value": "deadbeef",
          "deadline": 1614540154
        }    
      }
  }
  ```
  Output 2:
  ```json
  {
      "workflow": "agglo",
      "cicdVcsHash": "deadbeef",
      "state": "passed",
      "internal:completion:PipelineName": {
        "completionName": {
          "state": "complete",
          "resolved": {
            "head": true,
            "cicdVcsHash": true
          },
          "value": "deadbeef",
          "deadline": 1614540154
        }    
      }
  }
  ```
 
- **Filter**: Filter (or inverse filter) fields based on string or regex match of key
 
  Example: Filter out keys matching the regex `foo.*`

  Input:
  ```json
  {
      "fooBar": 1,
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 1
      }
  }
  ```
  Output:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 1
      }
  }
  ```
- **Spawner**: Conditionally spawn a process

  _Spawn Types_: Local executable or script

  Example: Asynchronously spawn `/bin/doSomething` when `someMap.someNum` is greater than 1
 
  Input:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 2
      }
  }
  ```
 
  Action: `echo '<input JSON>' | /bin/doSomething`
 
  Output:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 2
      }
  }
  ```
 
- **Transformation**: Transform one or more fields

  _Transformation Types_: Copy, Map, MapRegex, MapAdd, MapMultiply, Count, LeftFold, RightFold

- **Tee**: Send the input map to an external key-value store, object store, file system or REST endpoint

  _Tee Types_: Key-value store, object store, local file, pub/sub, REST endpoint
 
  Example: Store copy of the input to a DynamoDB-backed key-value store.  Note that UUIDs are automatically
  generated and provided in the output map along with other non-secret connectivity information.
 
  Input:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 2
      }
  }
  ```
 
  Action: Store the input JSON in the configured Dynamo table.
 
  Output:
  ```json
  {
    "name": "frank",
    "someMap": {
        "hi": "hello",
        "someNum": 2
    },
    "internal:tee:output": [
        {
          "connectionString": "dynamo:endpoint=https://123.45.67.8,region=us-west-2,tableName=kvStoreA,prefixLength=4",
          "outputType": "kvStore",
          "uuid": "7b1987ed-2030-4973-9f81-9955d3a25d86"
        }
      ]
  }
  ```
- **Continuation**: Conditionally continue or stop processing for this event.  There are cases
where a single binge process will have multiple pipelines defined for multiple event types and you
may not want to process all events through every pipeline.  A continuation allows you to conditionally stop processing.

  Example: Stop processing where `name` is `frank`
 
  Input:
  ```json
  {
      "name": "frank",
      "someMap": {
          "hi": "hello",
          "someNum": 2
      }
  }
  ```
 
  Action: The pipeline is gracefully cancelled.
 
  Output: None

- **Entwine**: Anchor the input map to an immutable, shared timeline
 
  Agglo comes packaged with an immutable timer, called a `ticker`, that
  appends a block at a specified time interval.  The ticker has an `Anchor` API
  that allows processes to entwine proofs of their own history with the ticker.
  This allows multiple independent processes to define a partial ordering without
  explicitly coordinating.  Each `Anchor` call must show an immutable chain of events
  and proof that the events proceeded the last known event for that process.  If the ticker
  successfully verifies the proof, it will associate its current time with the events in the
  proof.
 
  An entwine process will conditionally append an input map to a `substream` and will periodically
  anchor the substream to a specified ticker.
 
  Currently, the `ticker` is a standalone server that does not rely on any external coordination, but
  can easily be extended to be backed by more complex immutable ledgers that utilize proof-of-work, proof-of-stake
  or PAXOS-like consensus.
 
  See the [Entwine](#entwine-immutable-partial-ordering-of-events) section for more details.
