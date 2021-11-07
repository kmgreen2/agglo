{
  "partitionUuid": "9eb9f91d-b8e6-461d-896d-5e553ade52f7",
  "pipelines": [
    {
      "name": "Binge",
      "processes": [
        {
          "name": "Entwine",
          "retryStrategy": {
            "numRetries": 0,
            "initialBackOffMs": 0
          },
          "instrumentation": {
            "enableTracing": false,
            "latency": false,
            "counter": false
          }
        },
        {
          "name": "PopLastEntwineOutput",
          "retryStrategy": {
            "numRetries": 0,
            "initialBackOffMs": 0
          },
          "instrumentation": {
            "enableTracing": false,
            "latency": false,
            "counter": false
          }
        },
        {
          "name": "MapEntwineOutput",
          "retryStrategy": {
            "numRetries": 0,
            "initialBackOffMs": 0
          },
          "instrumentation": {
            "enableTracing": false,
            "latency": false,
            "counter": false
          }
        },
        {
          "name": "SearchIndex",
          "retryStrategy": {
            "numRetries": 0,
            "initialBackOffMs": 0
          },
          "instrumentation": {
            "enableTracing": false,
            "latency": false,
            "counter": false
          }
        },
        {
          "name": "TeeOutput",
          "retryStrategy": {
            "numRetries": 0,
            "initialBackOffMs": 0
          },
          "instrumentation": {
            "enableTracing": false,
            "latency": false,
            "counter": false
          }
        }
      ],
      "enableTracing": false,
      "enableMetrics": false
    }
  ],
  "externalSystems": [
    {
      "name": "KVStore",
      "connectionString": "dynamo:endpoint=http://localhost:8000,region=us-west-2,tableName=kvStore,prefixLength=4",
      "externalType": "ExternalKVStore"
    },
    {
      "name": "ObjectStore",
      "connectionString": "s3:endpoint=localhost:9000,bucketName=localtest",
      "externalType": "ExternalObjectStore"
    },
    {
      "name": "SearchIndex",
      "connectionString": "elastic:index=test-index,nodes=http://localhost:9200,authString=basic;;",
      "externalType": "ExternalSearchIndex"
    },
    {
      "name": "LocalFileFor",
      "connectionString": "/tmp/binge-out",
      "externalType": "ExternalLocalFile"
    }
  ],
  "processDefinitions": [
    {
      "entwine": {
        "name": "Entwine",
        "streamStateStore": "KVStore",
        "objectStore": "ObjectStore",
        "pemPath": "/tmp/binge-SUBSTREAM_ID.pem",
        "subStreamID": "SUBSTREAM_ID",
        "tickerEndpoint": "localhost:8001",
        "tickerInterval": 1,
        "condition": {}
      }
    },
    {
      "tee": {
        "name": "SearchIndex",
        "condition": {},
        "outputConnectorRef": "SearchIndex",
        "transformerRef": "",
        "additionalBody": {}
      }
    },
    {
      "tee": {
        "name": "TeeOutput",
        "condition": {},
        "outputConnectorRef": "LocalFileFor",
        "transformerRef": "",
        "additionalBody": {}
      }
    },
    {
      "transformer": {
        "name": "PopLastEntwineOutput",
        "specs": [
          {
            "sourceField": "internal:entwine:output",
            "targetField": "entwineOutput",
            "transformation": {
              "transformationType": "TransformPopTail",
              "condition": {}
            }
          }
        ],
        "forwardInputFields": true
      }
    },
    {
      "transformer": {
        "name": "MapEntwineOutput",
        "specs": [
          {
            "sourceField": "entwineOutput.entwineUuid",
            "targetField": "entwineUuid",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          },
          {
            "sourceField": "entwineOutput.objectDescriptor.objectKey",
            "targetField": "entwinedObjectKey",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          },
          {
            "sourceField": "entwineOutput.objectDescriptor.objectStoreConnectionString",
            "targetField": "entwinedObjectDescriptor",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          },
          {
            "sourceField": "entwineOutput.subStreamID",
            "targetField": "entwinedSubStreamID",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          },
          {
            "sourceField": "entwineOutput.subStreamIndex",
            "targetField": "entwinedSubstreamIndex",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          },
          {
            "sourceField": "entwineOutput.tickerUuid",
            "targetField": "entwinedTickerUuid",
            "transformation": {
              "transformationType": "TransformCopy",
              "condition": {}
            }
          }
        ],
        "forwardInputFields": true
      }
    }
  ]
}
