# agglo: A Framework for Lightweight, Flexible Event Stream Processing

![Workflow Status](https://github.com/kmgreen2/agglo/workflows/Go/badge.svg?branch=main)

## Overview

Agglo is an experimental event stream processing framework that enables
lightweight, reliable and scalable stream processing without the use of or
alongside persistent object storage, key-value storage or stream processing
platforms.

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

## Design and Architecture

### Pipelines

### External System Support

## Entwine: Immutable, Partial Ordering of Events


