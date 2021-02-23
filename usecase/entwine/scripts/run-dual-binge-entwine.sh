ROOTDIR=$( dirname $0 )/../../..

function cleanup {
    docker kill dynamodb
    docker rm dynamodb
    docker kill minio
    docker rm minio
}

trap cleanup EXIT INT TERM

export AWS_ACCESS_KEY_ID=localtest
export AWS_SECRET_ACCESS_KEY=localtest
export AWS_DEFAULT_REGION=us-west-2

docker run -d -p 8000:8000 --name dynamodb amazon/dynamodb-local

sleep 2

# Create tables 
aws dynamodb create-table --endpoint-url http://localhost:8000 --region us-west-2 \
     --table-name kvStoreA \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

aws dynamodb create-table --endpoint-url http://localhost:8000 --region us-west-2 \
     --table-name kvStoreB \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

aws dynamodb create-table --endpoint-url http://localhost:8000 --region us-west-2 \
     --table-name tickerKVStore \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1

docker run -d -p 9000:9000 \
  -e "MINIO_ROOT_USER=localtest" \
  -e "MINIO_ROOT_PASSWORD=localtest" \
  --name minio \
  minio/minio server /data

sleep 2

mc alias set localtest http://localhost:9000 localtest localtest

# Create buckets to use in the test
mc mb localtestA
mc mb localtestB

# Generate key pairs

rm /tmp/a.pem*; ssh-keygen -f /tmp/a.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/a.pem.pub -e -m pem > /tmp/authenticators/A.pem
rm /tmp/b.pem*; ssh-keygen -f /tmp/b.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/b.pem.pub -e -m pem > /tmp/authenticators/B.pem
rm /tmp/ticker.pem*; ssh-keygen -f /tmp/ticker.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/ticker.pem.pub -e -m pem > /tmp/authenticators/ticker.pem


${ROOTDIR}/bin/ticker  --grpcPort 8001 -kvConnectionString "dynamo:endpoint=localhost:8000,region=us-west-2,tableName=tickerKVStore,prefixLength=4" \
    -authenticatorPath /tmp/authenticators -privateKeyPath /tmp/ticker.pem &
TICKER_PID=$!

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort 8008 -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${ROOTDIR}/usecase/entwine/config/binge-a.json &
BINGE1_PID=$!

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort 8009 -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${ROOTDIR}/usecase/entwine/config/binge-b.json &
BINGE2_PID=$!

sleep 2

kill -9 ${TICKER_PID} ${BINGE1_PID} ${BINGE2_PID}
