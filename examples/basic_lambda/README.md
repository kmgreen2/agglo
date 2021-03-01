# The Basic Lambda Example

This is the same as the [Basic Example](../basic), but relies on a locally managed
Lambda function, instead of a daemon for event processing.

## The Pipeline Configuration

See [Basic Example](../basic) for an explaination of the pipeline configuration and sample events.

## Build a Lambda Container and Run It

Amazon has a toolkit that allows you to run Lambdas locally using Docker.  We have a special make
target that will build a locally runnable Lambda function.

To build the local Lambda:

```
# make -C ../.. CONFIGFILE="examples/basic_lambda/basic_pipeline.json" LAMBDANAME="example" lambda-local
```

This will build an image tagged as `binge-lambda-example-local:latest`.

To invoke the local Lambda handler, run:

```
# docker run -p9000:8080 binge-lambda-example-local:latest
time="2021-03-01T23:23:33.224" level=info msg="exec '/binge' (cwd=/var/task, handler=lambda)"
```

You should see the above log message, which means the Lambda can be triggered.

## Trigger the Lambda with Some Events

As with the [Basic Example](../basic), create the directories needed for the `Tee` processes:

```
mkdir -p /tmp/before-events
mkdir -p /tmp/after-events
``` 

These directories will contain the original event payloads and the payloads
after being processed by the pipeline.

Trigger the Lambda function using the `genevents` tool:

```
let i=0
while (( ${i} < 10 )); do
    ../../bin/genevents  -silent -numEvents 1 -numThreads 1 -schema ./basic_event_gen.json -output "" | curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations"  -d@-
    let i=${i}+1
done
```

This will trigger the lambda 10 times using the provided event generator config.

The Lambda output of should have 3 lines per invocation:

```
...
START RequestId: 5d3fc3ea-fe6b-4a35-82b5-75c16b1f2b3f Version: $LATEST
END RequestId: 5d3fc3ea-fe6b-4a35-82b5-75c16b1f2b3f
REPORT RequestId: 5d3fc3ea-fe6b-4a35-82b5-75c16b1f2b3f	Duration: 2.63 ms	Billed Duration: 100 ms	Memory Size: 3008 MB	Max Memory Used: 3008 MB
...
```

For information on how to analyze the events, see [Basic Example](../basic).
