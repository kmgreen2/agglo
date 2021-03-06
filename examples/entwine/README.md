# Entwine Example

For this example, we will create two binge processes that will append messages to their own immutable substreams.  We
will also start a ticker processes that exposes an API for substreams to anchor to.  Once getting all of the processes
up, we will append a few messages to each substream and show the partial order maintained using the ticker.

## Bring Up the Dependencies

Run the following:

```
# ./setup-dynamo-minio-ticker-binge.sh
```

You will see a lot of output.  Here is what is happening:

1. Start local DynamoDB instance, which is used to persist the substreams and ticker
2. Start a local Minio instance, which is used to store the content associated with the substream messages
3. Create keypairs for each binge process and the ticker process (used to sign/verify signatures)
3. Start a ticker
4. Send a tick message to the ticker (the tick API forces a new ticker message)
5. Start two binge daemon processes (8008 and 8009), which will register with the ticker on startup.  Each process is
 configured to anchor with the ticker every **2** messages

From here, we just have to tell the ticker to tick and send events to the binge processes.

## Create Events

Let's create an event per second, alternating the binge process to send the events to.  We will also tick before each
event is generated.

```
echo '{"idx":1, "aIdx": 1}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8008/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":2, "bIdx": 1}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8009/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":3, "aIdx": 2}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8008/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":4, "bIdx": 2}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8009/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":5, "aIdx": 3}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8008/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":6, "bIdx": 3}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8009/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":7, "aIdx": 4}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8008/entwine -o /dev/null
curl --silent -X POST http://localhost:8002/api/v1/tick
echo '{"idx":8, "bIdx": 4}' | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8009/entwine -o /dev/null
```

This will create a total of 8 messages, 4 in each substream.  You can see the events by inspecting `/tmp/a-out` and `/tmp/b-out`.

Here is a truncated listing of the messages for the first binge process:
```
# echo /tmp/a-out/* | xargs cat | jq
{
  "aIdx": 2,
  "bingeProcess": "A",
  "idx": 3,
  "internal:entwine:output": [
    {
      "entwineUuid": "0fc21052-05ac-4afa-a8bd-3dc57293eb99",
      "objectDescriptor": {
        "objectKey": "5691fe5e-50b8-45e5-9716-180ac16e2798",
        "objectStoreConnectionString": "s3:endpoint=localhost:9000,bucketName=localtesta"
      },
      "subStreamID": "e546e731-2b86-43cd-b847-764f437a7835",
      "tickerUuid": "0df58628-0ebf-40da-a29c-fe67d30067fa"
    }
  ],
  "internal:messageID": "16d210da-56bd-4cef-94d0-98333f9fd60e",
  "internal:name": "BingeA",
  "internal:partitionID": "9eb9f91d-b8e6-461d-896d-5e553ade52f7"
}
...
```
You can use `entwinectl` to verify the ordering of events.  Since we ticked the ticker between events and the substreams
anchor after every event, we know that the first event (idx:1, aIdx:1) should definitely happen before the second event
(idx: 2, bIdx: 1).

Get the first event:
```
$ echo /tmp/a-out/* | xargs cat | jq 'select(.idx==5)'
{
  "aIdx": 3,
  "bingeProcess": "A",
  "idx": 5,
  "internal:entwine:output": [
    {
      "entwineUuid": "39384886-650a-4e11-9359-d673d2be5edb",
      "objectDescriptor": {
        "objectKey": "eeb2d56c-4222-4990-94c2-f3a8b115ef76",
        "objectStoreConnectionString": "s3:endpoint=localhost:9000,bucketName=localtesta"
      },
      "subStreamID": "e546e731-2b86-43cd-b847-764f437a7835",
      "tickerUuid": "308627c6-d24b-4fe6-a0fc-404ce913efd8"
    }
  ],
  "internal:messageID": "9b29ea7a-6961-48cf-aea3-8d2e12cd1c7e",
  "internal:name": "BingeA",
  "internal:partitionID": "9eb9f91d-b8e6-461d-896d-5e553ade52f7"
}
``` 

Get the second event:

```
$ echo /tmp/b-out/* | xargs cat | jq 'select(.idx==8)'
{
  "bIdx": 4,
  "bingeProcess": "B",
  "idx": 8,
  "internal:entwine:output": [
    {
      "entwineUuid": "1a50e0e0-1a58-418d-80a4-89e036efe1eb",
      "objectDescriptor": {
        "objectKey": "8b7df578-5f32-4ae0-8267-2a3fd9a430c7",
        "objectStoreConnectionString": "s3:endpoint=localhost:9000,bucketName=localtestb"
      },
      "subStreamID": "ed94a517-f030-4aa6-b390-6283ad24cab3",
      "tickerUuid": "308627c6-d24b-4fe6-a0fc-404ce913efd8"
    }
  ],
  "internal:messageID": "a9a256b3-0dda-4a9e-919d-e55743f45be1",
  "internal:name": "BingeB",
  "internal:partitionID": "9eb9f91d-b8e6-461d-896d-5e553ade52f7"
}
```

Use `entwinectl` to validate that `39384886-650a-4e11-9359-d673d2be5edb` happened before `1a50e0e0-1a58-418d-80a4-89e036efe1eb`

```
# The format for happenedBefore is <subStreamUuid>:<UUID>:<subStreamIdx>
 ../../bin/entwinectl  -tickerEndpoint localhost:8001 -command HappenedBefore \
    e546e731-2b86-43cd-b847-764f437a7835:39384886-650a-4e11-9359-d673d2be5edb:3 \
    ed94a517-f030-4aa6-b390-6283ad24cab3:37443ed1-5349-4cd3-a511-ac5fb70bd0ae:4

39384886-650a-4e11-9359-d673d2be5edb happenedBefore 37443ed1-5349-4cd3-a511-ac5fb70bd0ae: true
```
