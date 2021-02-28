# Agglo: A Framework for Lightweight, Flexible Event Stream Processing

![Workflow Status](https://github.com/kmgreen2/agglo/workflows/Go/badge.svg?branch=main)

## Overview

Agglo is an experimental event stream processing framework that enables
lightweight, reliable and scalable stream processing alongside persistent
object storage, key-value storage or stream processing platforms.

Binge (BINary-at-the-edGE) is the main artifact of the framework.  As the name
implies, it is a single, compiled binary and can run as a stateless daemon,
persistent daemon or as a stand-alone command.  This allows the same binary to
be deployed in edge gateways, as Kubernetes deployments, in load balancers,
cloud (lambda) functions, or anywhere else you need to perform stream
processing.  The deployed artifact is simply a binary and a single JSON config
that defines the event processing pipelines and configuration for connecting to
external object stores, key-value stores, HTTP/RPC endpoints and stream
processing systems.

With the exception of persistent daemon mode, binge is completely stateless and
can be seamlessly scaled, killed and restarted.  When run as a persistent
daemon, binge requires a volume (local or cloud managed, such as EBS) to
maintain a durable queue.  If a cloud-managed block storage volume is used to
back the durable queue, a persistent daemon may also be treated as if it was
stateless.  In this way, we like to think of binge as
nginx-for-stream-processing.

## Rationale and Approach

We already have plenty of stable, reliable stream processing frameworks, so
why do we need this one?

Most stream processing is done explicitly using a framework, such as Spark,
Flink, Kafka Streams, etc. or via a custom microservice architecture that
leverages a reliable queue or event bus.  Many stream processing tasks are
simple transformations, filters, aggregations, annotations, etc. that can be
done at the point of ingestion.  This becomes more and more important with IoT
use cases, where ingestion points (the edge) is far away from your operations
VPC(s).  In short, we believe that many stream processing tasks can be
accomplished without the use of stream processing systems and/or custom
microservice architectures.

Our approach is to provide a single multi-use binary that can handle most
stream processing tasks, where any stateful interactions are handled by
external key-value stores, object stores or file systems, and more complex
stream processing tasks can be forwarded to the appropriate stream processing
system.  This provides the flexibility to deploy the binary to the most
appropriate "edge" (e.g., load balancer, IoT gateway, Lambda function, etc.)
depending on the use case.  In addition, it provides the flexibility to rely on
external systems for stateful interactions, which can also exist at the edge
(e.g., Agglo provides in-memory key-value stores), existing on-prem deployment
or cloud-managed deployment.

Agglo/binge is not intended to replace existing stream processing systems.
While there are many cases where it could be the main technology used for
stream processing, it is also a perfect fit for complimenting existing stream
processing workloads, since it can preprocess and route events from the edge
to a centrally managed stream processing system or event bus.

The main insight here is that there is no need to do all stream processing in
traditional SPE.  The value here is to provide the ability to tradeoff cost,
latency, throughput, consistency by using binge with other systems.  Our
hypothesis is that binge can be used as the basis of any event-driven
architecture, where the binge processes are distributed to different edges (IoT
gateways, LB, Kubernetes, Lambda) and utilize managed systems for persistence
and more involved processing.

While we believe that Agglo/binge is best used in conjunction with existing
stream processing systems, we also believe that it can be the main building
block for IoT event processing, building complex iPaaS
(integration-platform-as-a-service) architectures and any future use case that
requires highly distributed event processing from varied sources.

ToDo: Add some pictures with a few examples.

## Getting Started

To make life easier, it is best to make sure the tests pass before doing anything else.  This requires that the
following is installed:

- [aws-cli](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html): Local AWS services are used for testing
- [go-mock](https://github.com/golang/mock): Mocks are generated at test time (I'll probably just check the mocks in eventually)
- [minio-client](https://docs.min.io/docs/minio-client-quickstart-guide.html): Used for testing local object stores

First, make sure everything builds and the unit-tests pass:

```
make test
```

Once the tests pass, you can try out some of the [Examples](./examples) to get an idea of how
agglo and binge work.

## Configuration Generator

The raw JSON pipeline configs can become very unwieldy.  Please use the [Agglo Config Generator](./web/config-builder)
to generate all of you configs.  A few caveats:

1. The maintainer (@kmgreen2) is not a front-end developer.  I tried to do things the "React-way", but am sure things can be way better.
2. It does not explicitly save your progress, so you should periodically "Download" from the "Editor" screen to ensure you do not lose you edits.

This will be enhanced as I find time or others to help.  This was my first React project, so am sure someone else can do way better.

## Design and Architecture

There are 5 main components to Agglo:

1. Process: A stage in the pipeline that will consume an input map, perform an operation (annotation, aggregation, completion, filter, spawner, transformation, tee, continuation, entwine) and output a map.
2. Pipeline: An ordered collection of processes applied to an event.
3. External systems: A connector to an external system that can used by a process.  Today, we support S3-like Object Stores, POSIX filesystems, Key-value stores, messaging/pubsub systems and REST endpoints.
4. Binge: A event processing binary and can run as a stateless daemon, persistent daemon or as a stand-alone command.
5. Pipeline configuration: A binge instance is instantiated using the pipeline configuration that contains one or more pipelines and their dependent processes and external systems.

### Processes

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

### Pipelines

Please refer to the [examples](https://github.com/kmgreen2/agglo/tree/main/examples) for
examples of pipelines.

### External Systems

  Adding a new external system is as "simple" as implementing the proper interface.

- **KVStores**
 
  _Current Support_: In-memory local key-value store, local DynamoDB and managed DynamoDB
 
  See [KVStores](https://github.com/kmgreen2/agglo/tree/main/pkg/kvs) for the current
  implementations.

  Implementing a new key-value store is a matter of implementing the following interface:
  ```
  type KVStore interface {
  	AtomicPut(ctx context.Context, key string, prev, value []byte) error
  	AtomicDelete(ctx context.Context, key string, prev []byte) error
  	Put(ctx context.Context, key string, value []byte) error
  	Get(ctx context.Context, key string) ([]byte, error)
  	Head(ctx context.Context, key string) error
  	Delete(ctx context.Context, key string) error
  	List(ctx context.Context, prefix string) ([]string, error)
  	ConnectionString() string
  	Close() error
  }
  ```

- **Object Stores**

  _Current Support_: In-memory local object store and Minio (supports most managed object stores).
 
  See [ObjectStores](https://github.com/kmgreen2/agglo/tree/main/pkg/storage) for the current
  implementations.

  Implementing a new object store is a matter of implementing the following interface:

  ```
  type ObjectStore interface {
  	Put(ctx context.Context, key string, reader io.Reader)	error
  	Get(ctx context.Context, key string) (io.Reader, error)
  	Head(ctx context.Context, key string) error
  	Delete(ctx context.Context, key string) error
  	List(ctx context.Context, prefix string) ([]string, error)
  	ConnectionString() string
  	ObjectStoreBackendParams() ObjectStoreBackendParams
  }
  ```

- **Pubsub**

  _Current Support_: In-memory local pub/sub (Kafka coming soon)
 
  See [Streaming](https://github.com/kmgreen2/agglo/tree/main/pkg/streaming) for the current
  implementations.

  Implementing a new publisher is a matter of implementing the following interface:
  ```
  type Publisher interface {
  	Publish(ctx context.Context, b []byte) error
  	Flush(ctx context.Context, timeout time.Duration) error
  	Close() error
  	ConnectionString() string
  }
  ```

- **REST Endpoints**

  All HTTP/S requests are supported.

## Entwine: Immutable, Partial Ordering of Events

Entwine is a special process that provides the ability to "entwine" event
timelines.  It is similar to existing blockchain technologies in that is relies
on hash chains for immutability and zero-knowledge proofs for secure
verification of individual event timelines.  There is no consensus mechanism.
The assumption is that each event timelines is generated by an individual,
single organization, binge binary or collection of binge binaries.

An entwined stream of events is a collection of independent substreams that
anchor to a ticker stream.  Currently, the ticker stream is a simple hash chain
that "ticks" at an interval and allows substreams to anchor.  Each anchor
request must contain a signed proof that must be consistent with the ticker's
view of that substream's history.  If verified, the anchor request succeeds.

The `HappenedBefore` relation can be applied to all events on an entwined
stream.  This means that we can define a partial ordering of all events in the
stream, where the individual events are independently managed.

A basic example of this could be binge agents deployed to every news outlet in
the world.  Each agent will process every news story from each outlet, where
each outlet will maintain a local ordering in their own private substream.  As
long as each substream anchors to the ticker stream at regular intervals, we
can determine the partial order in which stories are broken and maintain a
sort-of truth index for new outlets (e.g. you cannot change history).

