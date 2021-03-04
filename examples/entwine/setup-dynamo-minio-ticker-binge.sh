ROOTDIR=$( dirname $0 )/../..
THISDIR=$( dirname $0 )

BINGEA_PORT=8008
BINGEB_PORT=8009
TICKER_PORT=8002
DYNAMO_PORT=8000
MINIO_PORT=9000

SETUP_DYNAMO=
SETUP_MINIO=
SETUP_DONE=

function cleanup {
    if [[ -z ${SETUP_DONE} ]]; then
        if [[ -n ${SETUP_DYNAMO} ]]; then
            docker kill dynamodb
            docker rm dynamodb
        fi
        if [[ -n ${SETUP_MINIO} ]]; then
            docker kill minio
            docker rm minio
        fi
    fi
}

trap cleanup EXIT INT TERM

export AWS_ACCESS_KEY_ID=localtest
export AWS_SECRET_ACCESS_KEY=localtest
export AWS_DEFAULT_REGION=us-west-2

mkdir -p /tmp/authenticators
mkdir -p /tmp/a-out
mkdir -p /tmp/b-out

rm /tmp/a-out/*
rm /tmp/b-out/*

docker run -d -p ${DYNAMO_PORT}:${DYNAMO_PORT} --name dynamodb amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb
sleep 2

# Create tables 
aws dynamodb create-table --endpoint-url http://localhost:${DYNAMO_PORT} --region us-west-2 \
     --table-name kvStoreA \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 || exit 1

aws dynamodb create-table --endpoint-url http://localhost:${DYNAMO_PORT} --region us-west-2 \
     --table-name kvStoreB \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 || exit 1

aws dynamodb create-table --endpoint-url http://localhost:${DYNAMO_PORT} --region us-west-2 \
     --table-name tickerKVStore \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1 || exit 1

SETUP_DYNAMO=1

docker run -d -p ${MINIO_PORT}:${MINIO_PORT} \
    -e "MINIO_ROOT_USER=localtest" \
    -e "MINIO_ROOT_PASSWORD=localtest" \
    --name minio \
    minio/minio server /data || exit 1

sleep 2
mc alias set localtest http://localhost:9000 localtest localtest

SETUP_MINIO=1

# Create buckets to use in the test
mc mb localtest/localtesta || exit 1
mc mb localtest/localtestb || exit 1

# Generate key pairs

rm /tmp/a.pem*; ssh-keygen -f /tmp/a.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/a.pem.pub -e -m pem > /tmp/authenticators/e546e731-2b86-43cd-b847-764f437a7835.pem || exit 1
rm /tmp/b.pem*; ssh-keygen -f /tmp/b.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/b.pem.pub -e -m pem > /tmp/authenticators/ed94a517-f030-4aa6-b390-6283ad24cab3.pem || exit 1
rm /tmp/ticker.pem*; ssh-keygen -f /tmp/ticker.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/ticker.pem.pub -e -m pem > /tmp/authenticators/ticker.pem || exit 1

${ROOTDIR}/bin/ticker  --grpcPort 8001 --httpPort ${TICKER_PORT} -kvConnectionString "dynamo:endpoint=http://localhost:${DYNAMO_PORT},region=us-west-2,tableName=tickerKVStore,prefixLength=4" \
        -authenticatorPath /tmp/authenticators -privateKeyPath /tmp/ticker.pem > /tmp/ticker.out 2>&1 &
TICKER_PID=$!

# Wait for the ticker to come up
sleep 5

curl --silent -X POST http://localhost:${TICKER_PORT}/api/v1/tick > /dev/null 2>&1

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort ${BINGEA_PORT} -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${THISDIR}/binge-a.json > /tmp/binge-a.out 2>&1 &
BINGE1_PID=$!

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort ${BINGEB_PORT} -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${THISDIR}/binge-b.json > /tmp/binge-b.out 2>&1 &
BINGE2_PID=$!

echo "All services are up!"
echo "Run the following to cleanup: kill -9 ${TICKER_PID} ${BINGE1_PID} ${BINGE2_PID}; ${THISDIR}/stop-dynamo-minio-ticker-binge.sh"

SETUP_DONE=1
