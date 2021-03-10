# Text Tokenization and Lemmatization Example

This is an example of doing text pre-processing with binge.  There are two main components:

1. Pipeline that will store the oginal payload in S3, spawns a tokenizer that tokenizes a payload and forwards the tokens to a REST endpoint.
2. Dumbserver: a simple REST endpoint that prints payloads posted to it.

This illustrates how binge can be deployed to many edge locations to persist (store to S3) and pre-process (tokenize).

## The Pipeline Configuration

There are only 3 processes and one pipeline here:

```
cat nlp-pipeline.json | jq '.pipelines[].processes[].name'
"Store Original in S3"
"Tokenize"
"Send Tokens to Token Endpoint"
```

The pipeline will immediately store the inbound event in S3, then run a tokenize process and send the tokens to a REST endpoint.

Here are the external systems references in the process definitions (local dumserver endopint and S3 endpoint):

```
[
  {
    "name": "TokenEndpoint",
    "connectionString": "http://localhost:1337/webhook",
    "externalType": "ExternalHttp"
  },
  {
    "name": "S3ObjectStore",
    "connectionString": "s3:endpoint=localhost:9000,bucketName=localtest",
    "externalType": "ExternalObjectStore"
  }
]
```

Here are the process definitions:

```
$ cat nlp-pipeline.json | jq '.processDefinitions'
[
  {
    "spawner": {
      "name": "Tokenize",
      "condition": {
        "exists": {
          "ops": [
            {
              "key": "text",
              "op": "Exists"
            }
          ]
        }
      },
      "delayInMs": 0,
      "doSync": true,
      "job": {
        "runnable": {
          "pathToExec": "/usr/local/bin/docker",
          "cmdArgs": [
            "run",
            "-i",
            "tokenizer",
            "python3",
            "main.py",
            "--tokenize_type",
            "regular"
          ]
        }
      }
    }
  },
  {
    "tee": {
      "name": "Store Original in S3",
      "condition": {},
      "outputConnectorRef": "S3ObjectStore",
      "transformerRef": "",
      "additionalBody": {}
    }
  },
  {
    "tee": {
      "name": "Send Tokens to Token Endpoint",
      "condition": {},
      "outputConnectorRef": "TokenEndpoint",
      "transformerRef": "",
      "additionalBody": {}
    }
  }
]
```

The two `Tee` processes are very straightforward.  The `Spawn` process will spawn a Docker container (see `./Dockerfile`) containing 
the tokenizer.  The tokenizer will tokenize any text under teh `text` key and forward the output to the next process.

## Generate Some Example Events

To get an idea of the events this example sends, run: 
```
../../bin/genevents  -numEvents 2 -numThreads 1 -schema ./nlp-pipeline-gen.json -output ""
{
	"someNum": 946.4581792405112,
	"someString": "b39ea4e21e6436",
	"text": "I'm heading back to Colorado tomorrow after being down in Santa Barbara over the weekend for the festival there. I will be making October plans once there and will try to arrange so I'm back here for the birthday if possible. I'll let you know as soon as I know the doctor's appointment schedule and my flight plans."
}

{
	"someNum": 296.6977473334879,
	"someString": "0f5f3ce636066",
	"text": "He looked at the sand. Picking up a handful, he wondered how many grains were in his hand. Hundreds of thousands? 'Not enough,' the said under his breath. I need more."
}
```

Here, the `text` payloads will be tokenized and forwarded to the REST `Tee` processes.

## Run the Example

The entire example is contained in `./run.sh`:

```
# ./run.sh 
##### There will be a ton of output ######
```

If everything goes well, you should eventually see an example event's S3 object and dumbserver payload:

```
Example Message ID: "5f0c58a0-c348-426b-947f-7780f3c9bafd"
Example S3 object:
{
  "internal:messageID": "5f0c58a0-c348-426b-947f-7780f3c9bafd",
  "internal:name": "Tokenizer",
  "internal:partitionID": "9926252d-d92a-48e4-89d3-74b1b47892dc",
  "someNum": 946.4581792405112,
  "someString": "b39ea4e21e6436",
  "text": "I'm heading back to Colorado tomorrow after being down in Santa Barbara over the weekend for the festival there. I will be making October plans once there and will try to arrange so I'm back here for the birthday if possible. I'll let you know as soon as I know the doctor's appointment schedule and my flight plans."
}
Payload from Dumbserver:
{
  "internal:messageID": "08195e48-59a3-4d78-b793-4a590ab5102d",
  "internal:name": "Tokenizer",
  "internal:partitionID": "9926252d-d92a-48e4-89d3-74b1b47892dc",
  "internal:spawn:output": [
    {
      "Tokenize": {
        "tokens": [
          "patiently",
          "waited",
          "number",
          "called",
          ".",
          "desire",
          ",",
          "mom",
          "insisted",
          ".",
          "'s",
          "resisted",
          "first",
          ",",
          "realized",
          "simply",
          "easier",
          "appease",
          ".",
          "mom",
          "tended",
          "way",
          ".",
          "would",
          "keep",
          "insisting",
          "wore",
          "wanted",
          ".",
          ",",
          ",",
          "patiently",
          "waiting",
          "number",
          "called",
          "."
        ]
      }
    }
  ],
  "internal:tee:output": [
    {
      "connectionString": "s3:endpoint=localhost:9000,bucketName=localtest",
      "outputType": "object",
      "uuid": "1aeba880-12c1-4c57-85ef-a78779aa803b"
    }
  ],
  "someNum": 217.3724615564126,
  "someString": "ed8be22a024786c72",
  "text": "She patiently waited for his number to be called. She had no desire to be there, but her mom had insisted that she go. She's resisted at first, but over time she realized it was simply easier to appease her and go. Mom tended to be that way. She would keep insisting until you wore down and did what she wanted. So, here she sat, patiently waiting for her number to be called."
}
```

