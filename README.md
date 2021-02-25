# agglo: A Framework for Lightweight, Flexible Event Stream Processing

![Workflow Status](https://github.com/kmgreen2/agglo/workflows/Go/badge.svg?branch=main)

## Overview

Agglo is an experimental event stream processing framework that enables
lightweight, reliable and scalable stream processing alongside persistent
object storage, key-value storage or stream processing platforms.

Binge (BINary-at-the-edGE) is the main artifact of the framework.  As the name
implies, it is a single, compiled binary and can run as a stateless daemon,
persistent daemon or as a standalone command.  This allows the same binary to
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
why do we need this one?  There are two big reasons:

1. Most stream processing is done explicitly using a framework, such as Spark,
Flink, Kafka Streams, etc. or via a custom microservice architecture that
leverages a reliable queue or event bus.

2. Many stream processing tasks are simple transformations, filters,
aggregations, annotations, etc. that can be done at the point of ingestion.
This becomes more and more important with IoT use cases, where ingestion points
(the edge) is far away from your operations VPC(s).

3. Building bespoke micropservices for integration-based event processing is
overkill.

Our approach is to provide a single multi-use binary that can handle most
stream processing tasks, where any stateful interactions are handled by
exrternal key-value stores, object stores or file systems, and more complex
stream processing tasks can be forwarded to the appropriate stream processing
system.  For example, binge is a perfect fit for complimenting existing stream
processing workloads, since it can preprocess and route events at the edge.

Main insight: No need to do all stream processing in traditional SPE.  Can
tradeoff cost, latency, throughput, consistency by using binge with other
systems.  Hypothesis, can use binge as basis of any event-driven architecture,
where the binge processes are distributed to different edges (IoT gateways, LB,
Kubernetes, Lambda) and utilize managed systems for persistence and more
involved processing.

Draw a picture of this ^^^: Events coming into many binge instances, which are
relying on managed systems.


## Getting Started

To make life easier, it is best to make sure the tests pass before doing antyhing else.  This requires that the
following is installed:

- [aws-cli](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-install.html): Local AWS services are used for testing
- [go-mock](https://github.com/golang/mock): Mocks are generated at test time (I'll probably just check the mocks in eventually)
- [minio-client](https://docs.min.io/docs/minio-client-quickstart-guide.html): Used for testing local object stores

First, make sure eveything builds and the unit-tests pass:

```
$ make test
```

Once the tests pass, you can try out some of the [Examples](./examples) to get an idea of how
agglo and binge work.

## Configuration Generator

The raw JSON pipeline configs can become very unwieldy.  Please use the [Agglo Config Generator](https://github.com/kmgreen2/agglo-config-builder)
to generate all of you configs.  A few caveats:

1. The maintainer (@kmgreen2) is not a front-end developer.  I tried to do things the "React-way", but am sure things can be way better.
2. It does not explicitly save your progress, so you should periodically "Download" from the "Editor" screen to ensure you do not lose you edits.

This will be enhanced as I find time or others to help.  This was my first React project, so am sure someone else can do way better.

## Design and Architecture

There are 5 main components to Agglo:

1. Process: A stage in the pipeline that will consume an input map, perform an operation (annotation, aggregation, completion, filter, spawner, transformation, tee, continuation, entwine) and output a map.
2. Pipeline: An ordered collection of processes, 
3. External systems: A connector to an external system that can used by a process.  Today, we support S3-like Object Stores, POSIX filesystems, Key-value stores, messaging/pubsub systems and REST endpoints.
4. Binge: A event processing binary and can run as a stateless daemon, persistent daemon or as a standalone command.
5. Pipeline configuration: A binge instance is instantiated uwing the pipeline configuration that contains one or more pipelines and their dependent processes and external systems.

### Processes

- Annotation
- Aggregation
- Completion
- Filter
- Spawner
- Transformation
- Tee
- Continuation
- Entwine

### Pipelines

### External Systems

Adding a new external system is as "simple" as implementing the proper interface.

- KVStores: in-memory, Dynamo
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

- Object Stores: in-memory, S3-like

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

- Pubsub: in-memory

```
type Publisher interface {
	Publish(ctx context.Context, b []byte) error
	Flush(ctx context.Context, timeout time.Duration) error
	Close() error
	ConnectionString() string
}
```

- REST endpoints: webhooks, microservices, etc.

### Binge

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

