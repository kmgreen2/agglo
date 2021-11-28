ROOTDIR=$( dirname $0 )/../..
THISDIR=$( dirname $0 )

BINGE_PORT=8008
TICKER_PORT=8002
DYNAMO_PORT=8000
MINIO_PORT=9000

SUB_STREAMS="e546e731-2b86-43cd-b847-764f437a7835 e546e731-2b86-43cd-b847-764f437a7834 e546e731-2b86-43cd-b847-764f437a7833 e546e731-2b86-43cd-b847-764f437a7832 e546e731-2b86-43cd-b847-764f437a7831 e546e731-2b86-43cd-b847-764f437a7830 e546e731-2b86-43cd-b847-764f437a7829"
SOURCE_MATCH="w3-businessinsider w3-foxnews w3-huffingtonpost w3-latimes w3-new-yahoo w3-reuters w3-siouxcityjournal"
let BINGE_START_PORT=9101
CURR_DATE="2017-08-24"
END_DATE_EXCLUSIVE="2017-09-01"

SETUP_DYNAMO=
SETUP_MINIO=
SETUP_ELASTIC=
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
        if [[ -n ${SETUP_ELASTIC} ]]; then
            ${ROOTDIR}/deployments/local/elastic/stop-elastic.sh
        fi
    else
        rm ${THISDIR}/us-test-corpus-2017-08-17-2wks.csv
        rm ${THISDIR}/binge-*.json
        kill -9 ${TICK_TICKER_PID}
        kill -9 ${TICKER_PID} ${BINGE_PIDS}; ${THISDIR}/stop.sh
    fi
    exit

}

function generate_key_pair {
    SUBSTREAM_ID=$1
    rm -f /tmp/binge-${SUBSTREAM_ID}.pem; ssh-keygen -f /tmp/binge-${SUBSTREAM_ID}.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/binge-${SUBSTREAM_ID}.pem.pub -e -m pem > /tmp/authenticators/${SUBSTREAM_ID}.pem || exit 1
}

trap cleanup EXIT INT TERM

export AWS_ACCESS_KEY_ID=localtest
export AWS_SECRET_ACCESS_KEY=localtest
export AWS_DEFAULT_REGION=us-west-2

mkdir -p /tmp/authenticators
mkdir -p /tmp/binge-out

rm /tmp/binge-out/*

docker run -d -p ${DYNAMO_PORT}:${DYNAMO_PORT} --name dynamodb amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb
sleep 2

# Create tables 
aws dynamodb create-table --endpoint-url http://localhost:${DYNAMO_PORT} --region us-west-2 \
     --table-name kvStore \
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

${ROOTDIR}/deployments/local/elastic/start-elastic.sh

SETUP_ELASTIC=1

# Create buckets to use in the test
mc mb localtest/localtest || exit 1

# Generate key pairs

# rm /tmp/binge.pem*; ssh-keygen -f /tmp/binge.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/binge.pem.pub -e -m pem > /tmp/authenticators/e546e731-2b86-43cd-b847-764f437a7835.pem || exit 1
rm /tmp/ticker.pem*; ssh-keygen -f /tmp/ticker.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/ticker.pem.pub -e -m pem > /tmp/authenticators/ticker.pem || exit 1

for substream_id in ${SUB_STREAMS}; do
    generate_key_pair ${substream_id}
done

${ROOTDIR}/bin/ticker  --grpcPort 8001 --httpPort ${TICKER_PORT} -kvConnectionString "dynamo:endpoint=http://localhost:${DYNAMO_PORT},region=us-west-2,tableName=tickerKVStore,prefixLength=4" \
        -authenticatorPath /tmp/authenticators -privateKeyPath /tmp/ticker.pem > /tmp/ticker.out 2>&1 &
TICKER_PID=$!

# Wait for the ticker to come up
sleep 5

curl --silent -X POST http://localhost:${TICKER_PORT}/api/v1/tick > /dev/null 2>&1

BINGE_PIDS=""

let port=${BINGE_START_PORT}
for substream_id in ${SUB_STREAMS}; do
    sed "s/SUBSTREAM_ID/${substream_id}/" ${THISDIR}/binge.tpl > ${THISDIR}/binge-${port}.json
    ${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort ${port} -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${THISDIR}/binge-${port}.json > /tmp/binge-${port}.out 2>&1 &
    BINGE_PIDS="${BINGE_PIDS} $!"
    let port=${port}+1
done


read -p  "All services are up! Press any key to continue..."

SETUP_DONE=1

unzip ${THISDIR}/us-test-corpus-2017-08-17-2wks.zip

${THISDIR}/tick_ticker.sh &
TICK_TICKER_PID=$!


while [[ ${CURR_DATE} != ${END_DATE_EXCLUSIVE} ]]; do
    CLIENT_PIDS=""
    echo "Processing ${CURR_DATE}"
    let port=${BINGE_START_PORT}
    for source in ${SOURCE_MATCH}; do
        ${ROOTDIR}/bin/csv2json -concurrency 1 -outEndpoint http://localhost:${port}/entwine -csvFile ${THISDIR}/us-test-corpus-2017-08-17-2wks.csv -csvMap 0:created,1:source,2:url,3:text -fieldMatchers 0:${CURR_DATE},1:${source} &
        CLIENT_PIDS="${CLIENT_PIDS} $!"
        echo "Port: ${port}"
        let port=${port}+1
    done

    wait ${CLIENT_PIDS}

    CURR_DATE=`date  -j -f %Y-%m-%d -v+1d ${CURR_DATE} +'%Y-%m-%d'`
done

