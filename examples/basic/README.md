# The Basic Example

This example will get you comfortable with understanding how binge works.
Agglo comes packaged with a simple event generator, that will generate events
based on a provided schema.  The events can then be piped to a file or send to
an HTTP endpoint.

In this example, we will invoke a stateless daemon that is instantiated with
the pipelines config defined in `basic_pipeline.json`.  Then, we will generate
events to send to the daemon for processing using `genevents` with the event
schema defined in `basic_event_gen.json`.

## The Pipeline Configuration

This configuration contains one pipeline that has 5 total processes and 3 external
systems.

### External Systems

There are two local file and one KVStore.  These external systems can be referenced
by the processes for storing state.  In this case, `BeforeEvents` and `AfterEvents`
will be used by `Tee` processes to store event state at the beginning and end of
each pipeline run.  The `AggregatorStateKVStore` system is used for maintaining
aggregation state.


Here is the external system definition in the config:
```
$ cat basic_pipeline.json  | jq '.externalSystems'
[
  {
    "name": "BeforeEvents",
    "connectionString": "/tmp/before-events",
    "externalType": "ExternalLocalFile"
  },
  {
    "name": "AfterEvents",
    "connectionString": "/tmp/after-events",
    "externalType": "ExternalLocalFile"
  },
  {
    "name": "AggregatorStateKVStore",
    "connectionString": "mem:aggregatorState",
    "externalType": "ExternalKVStore"
  }
]
```

### Pipeline and Processes

This example only contains a single pipeline with 5 processes:

```
$ cat basic_pipeline.json  | jq '.pipelines[].processes[].name'
"BeforeTee"
"AnnotateFooBar"
"SumSomeNum"
"RemovePrefix"
"AfterTee"
```


- BeforeTee: Send input event payload to `BeforeEvents`

```
$ cat basic_pipeline.json  | jq '.processDefinitions[] | select(.tee.name=="BeforeTee")'
{
  "tee": {
    "name": "BeforeTee",
    "condition": {},
    "outputConnectorRef": "BeforeEvents",
    "transformerRef": "",
    "additionalBody": {}
  }
}
```

- AnnotateFooBar: Will add `{'foobar': 'found'}` to all events where `"someDict.someFixed" == "foobar"`

```
$ cat basic_pipeline.json  | jq '.processDefinitions[] | select(.annotator.name=="AnnotateFooBar")'
{
  "annotator": {
    "name": "AnnotateFooBar",
    "annotations": [
      {
        "fieldKey": "foobar",
        "value": "found",
        "condition": {
          "expression": {
            "comparator": {
              "lhs": {
                "literal": "someDict.someFixed"
              },
              "rhs": {
                "literal": "foobar"
              },
              "op": "Equal"
            }
          }
        }
      }
    ]
  }
}
```

- SumSomeNum: Maintain a sum of `someNum`, where the state is tracked in
  `AggregatorStateKVStore`.  The current aggregation state will be forwarded to
  the next process in the pipeline

```
$ cat basic_pipeline.json  | jq '.processDefinitions[] | select(.aggregator.name=="SumSomeNum")'
{
  "aggregator": {
    "name": "SumSomeNum",
    "condition": {},
    "stateStore": "AggregatorStateKVStore",
    "asyncCheckpoint": false,
    "forwardState": true,
    "aggregation": {
      "key": "someNum",
      "aggregationType": "AggSum",
      "groupByKeys": []
    }
  }
}

```

- RemovePrefix: Perform regex replacement of `someVocab` and map the result to `someVocabSansPrefix`

```
$ cat basic_pipeline.json  | jq '.processDefinitions[] | select(.transformer.name=="RemovePrefix")'
{
  "transformer": {
    "name": "RemovePrefix",
    "specs": [
      {
        "sourceField": "someVocab",
        "targetField": "someVocabSansPrefix",
        "transformation": {
          "transformationType": "TransformMapRegex",
          "condition": {},
          "mapRegexArgs": {
            "regex": "buzz/(.*)",
            "replace": "$1"
          }
        }
      }
    ],
    "forwardInputFields": false
  }
}
```

- AfterTee: Send input event payload to `AfterEvents`

```
$ cat basic_pipeline.json  | jq '.processDefinitions[] | select(.tee.name=="AfterTee")'
{
  "tee": {
    "name": "AfterTee",
    "condition": {},
    "outputConnectorRef": "AfterEvents",
    "transformerRef": "",
    "additionalBody": {}
  }
}
```

## Look at Some Sample Events

Before running the pipelines, it may be helpful to get a sense of the events that will be generated.  

Run the event generator with the provided event schema:

```
$ ../../bin/genevents  -numEvents 4 -numThreads 1 -schema ./basic_event_gen.json -output ""
{
	"someDict": {
		"someFixed": "foobar"
	},
	"someNum": 946.4581792405112,
	"someString": "db39ea4e21e64361"
}

{
	"someDict": {
		"someFixed": "foobar"
	},
	"someNum": 282.8681889825906,
	"someString": "0f5f3ce636066"
}

{
	"someDict": {
		"someFixed": "foobar"
	},
	"someNum": 635.3277379147563,
	"someString": "ee67e07d01"
}

{
	"someReference": 946.4581792405112,
	"someVocab": "buzz/buzz"
}

Total duration: 0 ms
RPS: 23405.226387
Thread 0: min: 0.000000 ms, max: 0.000000 ms, avg: 0.000000 ms
```

We specified `-output ""`, which send all events to `stdout`.  We will provide the binge endpoint, so
the events are processed by the instantiated daemon.

The binge daemon will process each event in the pipeline shown above.  Each
stage of the pipline will perform one or more actions and forward an updtated
event mapping to the next process in the pipeline.

## Run Binge and Send Some Events

First, create the directories needed for the `Tee` processes:

```
mkdir -p /tmp/before-events
mkdir -p /tmp/after-events
``` 

These directories will contain the original event payloads and the payloads
after being processed by the pipeline.

You will need two terminal windows.

In the first terminal window, run binge:
```
$ ../../bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 4 -runType stateless-daemon -exporter none -config ./basic_pipeline.json
```

In the second terminal window generate 16 events and send them to binge:

```
$ ../../bin/genevents -numEvents 16 -numThreads 1 -schema ./basic_event_gen.json -output http://localhost:80/binge
Total duration: 27 ms
RPS: 587.458165
Thread 0: min: 1.000000 ms, max: 2.000000 ms, avg: 1.625000 ms

```

The output of binge should look a bit like this (if you see much more than this, something probably went wrong):

```
{"level":"info","ts":1614214277.42994,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.432225,"caller":"server/daemon.go:147","msg":"Http handler: 1 ms"}
{"level":"info","ts":1614214277.4338748,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.435383,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.436975,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.4382899,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.439689,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.44118,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.442802,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.444543,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.446672,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.448403,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.4499981,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.451798,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.453253,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
{"level":"info","ts":1614214277.454529,"caller":"server/daemon.go:147","msg":"Http handler: 0 ms"}
```

Ok, so we have processed 16 events by or pipeline.  Let's look at an example of a single event that
has gone through the pipeline.

Select a random event from the local file store (remember, this is where the event payloads are written
in the first process of the pipeline):
```
$ cat /tmp/before-events/0ceb49f5-3bdb-40ba-a371-eeac6e06b87c.json | jq
{
  "internal:messageID": "afa89a9f-f4d7-41a0-80c5-5fbe299667da",
  "internal:name": "BasicPipeline",
  "internal:partitionID": "c5024ca8-d601-499f-bb24-daae74357b5e",
  "someDict": {
    "someFixed": "foobar"
  },
  "someNum": 367.3743072006624,
  "someString": "ce636066666ee67e0"
}
```

Notice that there are some fields with `inernal:` prefixes.  These are internal annotations added
to each message as they enter the pipeline.  Each event is given a unique UUID, which is provided
in the `internal:messageID` field.

We can use `internal:messageID` to get the final payload from `AfterEvents`:

```
$ echo /tmp/after-events/* | xargs grep -h afa89a9f-f4d7-41a0-80c5-5fbe299667da | jq
{
  "foobar": "found",
  "internal:aggregation:BasicPipeline": {
    "someNum:AggSum": {
      "value": 792.1585823718777
    }
  },
  "internal:messageID": "afa89a9f-f4d7-41a0-80c5-5fbe299667da",
  "internal:name": "BasicPipeline",
  "internal:partitionID": "c5024ca8-d601-499f-bb24-daae74357b5e",
  "internal:tee:output": [
    {
      "connectionString": "/tmp/before-events",
      "outputType": "localfile",
      "uuid": "0ceb49f5-3bdb-40ba-a371-eeac6e06b87c"
    }
  ],
  "someDict": {
    "someFixed": "foobar"
  },
  "someNum": 367.3743072006624,
  "someString": "ce636066666ee67e0"
}

```

Looks like a lot has happened in the pipeline!

- The current state of the summation aggregation is provided in the `internal:aggregation:BasicPipeline` section
- The BeforeEvents process has attached information about how the payload was stored in that process
- We see that the `AnnotatieFooBar` process has added `{"foobar": "found"}`, since `someDict.someFixed == "foobar"

