# Traffic Bifurcation

This is a simple example that shows how binge can be used to bifurcate and transform traffic for testing purposes.
Canaries are a common way to slowly test new changes.  Canarying a change usually involves slowly routing production traffic
to new instances.  If something goes wrong, you can simply re-reroute traffic.

Traffic bifurcation is a little different.  In this case, instead of routing production requests to new instances, we can
conditionally mirror/route requests, process them (perform anonymization, filter, transform, etc.) and forward them to a any
instance(s), even non-production instances.  That is, traffic bifurcation allows us to canary (do not mirror, only route) and
test non-production instances with real production traffic.

Here we give a very simple example of mirroring production requests to a test instance.

## The Pipeline Configuration

The [pipeline configuration](./traffic-bifurcate.json) only contains two processes: `Forward to Test` and 
`Forward to Production`.

These are simple `Tee` processes that references RESTful endpoints:

```
"externalSystems": [
    {
      "name": "ServiceA",
      "connectionString": "http://localhost:1337/webhook",
      "externalType": "ExternalHttp"
    },
    {
      "name": "ServiceB",
      "connectionString": "http://localhost:1338/webhook",
      "externalType": "ExternalHttp"
    }
  ],
```

In this case, "production" is `ServiceA`.  All requests will be routed to the production endpoint:

```
 {
      "tee": {
        "name": "Forward to Service Production",
        "condition": {},
        "outputConnectorRef": "ServiceA",
        "transformerRef": "",
        "additionalBody": {}
      }
    }
```

The "test" instance is `ServiceB`.  Only requests containing `{"doTest": "true"}` will be copied and
routed to the test instance.  In addition, `{"routedTest": true}` is added as `additionalBody` to the request.  
Finally, we define a simple transformation and will map `someNum` to `newNum` (defined by the `APIChange` transformation):

```
{
      "tee": {
        "name": "Forward to Test",
        "condition": {
          "expression": {
            "comparator": {
              "lhs": {
                "variable": {
                  "name": "doTest"
                }
              },
              "rhs": {
                "literal": "true"
              },
              "op": "Equal"
            }
          }
        },
        "outputConnectorRef": "ServiceB",
        "transformerRef": "APIChange",
        "additionalBody": {
          "routedTest": true
        }
      }
    }
```

## Running the Example

The example is self-contained and can be run by running `./run.sh`.  The script is very simple:
```
../../bin/dumbserver -port 1337 >/tmp/dumbserver.1337 2>&1 &
PID1=$!
../../bin/dumbserver -port 1338 >/tmp/dumbserver.1338 2>&1 &
PID2=$!
../../bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 4 -runType stateless-daemon -exporter none -config ./traffic-bifurcate.json >/tmp/binge.out 2>&1 &
PID3=$!

sleep 5

../../bin/genevents  -numEvents 20 -numThreads 1 -schema ./traffic-bifurcate-gen.json -output http://localhost:80/binge

kill -9 ${PID1} ${PID2} ${PID3}
```

Here is what is going on:

1. Starting two `dumbserver` instances to act as `ServiceA` and `ServiceB`.  `dumbserver` is a simple REST server
that will print all bodies sent to its `/webhook` endpoint.

2. Starting the `binge` processes with the [pipeline configuration](./traffic-bifurcate.json)

3. Generating events that will randomly choose to add `{"doTest": "true"}` or `{"doTest": "false"}`

We expect to see a subset of events go to out "test" instance (`ServiceB` or `dumbserver:1338`).

Run it:

```
# ./run.sh 
  Total duration: 24 ms
  RPS: 820.099289
  Thread 0: min: 0.000000 ms, max: 3.000000 ms, avg: 1.150000 ms
```

Checking the logs of `ServiceA`, we see all 20 requests get routed to "production":

```
# cat /tmp/dumbserver.1337 | jq '.doTest'
"false"
"false"
"true"
"false"
"false"
"false"
"false"
"true"
"false"
"true"
"false"
"true"
"true"
"false"
"false"
"false"
"true"
"true"
"true"
"false"

```

As expected `ServiceB` only processes requests that have `{"doTest": "true"}`:

```
# cat /tmp/dumbserver.1338 | jq '.doTest'
  "true"
  "true"
  "true"
  "true"
  "true"
  "true"
  "true"
  "true"
```

We also see that the additional `routedTest` entry is added to these requests:

```
# cat /tmp/dumbserver.1338 | jq '.routedTest'
true
true
true
true
true
true
true
true
```

Finally, the transformer has transformed `someNum` to `newNum`, but still copied `someNum` (`forwardInputFields` is `true`):

```
# cat /tmp/dumbserver.1338 | head -n 1 | jq  
{
  "doTest": "true",
  "internal:messageID": "97b3f3a0-8cb7-4240-a4c1-4f283e1401e7",
  "internal:name": "TrafficMirror",
  "internal:partitionID": "e3f9e001-4920-45da-b053-18588e45804b",
  "newNum": 722.8221286178008,
  "routedTest": true,
  "someDict": {
    "someFixed": "foobar"
  },
  "someNum": 722.8221286178008,
  "someString": "f5f3ce636066666"
}
```