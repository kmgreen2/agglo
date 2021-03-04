ROOTDIR=$( dirname $0 )/../../../..

function cleanup {
    if [[ -z ${CI_TEST} ]]; then
        docker kill dynamodb
        docker rm dynamodb
        docker kill minio
        docker rm minio
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

if [[ -z ${CI_TEST} ]]; then
    docker run -d -p 8000:8000 --name dynamodb amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb
    sleep 2
fi

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

if [[ -z ${CI_TEST} ]]; then
    docker run -d -p 9000:9000 \
        -e "MINIO_ROOT_USER=localtest" \
        -e "MINIO_ROOT_PASSWORD=localtest" \
        --name minio \
        minio/minio server /data

    sleep 2
    mc alias set localtest http://localhost:9000 localtest localtest
fi


# Create buckets to use in the test
mc mb localtest/localtesta
mc mb localtest/localtestb

# Generate key pairs

rm /tmp/a.pem*; ssh-keygen -f /tmp/a.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/a.pem.pub -e -m pem > /tmp/authenticators/e546e731-2b86-43cd-b847-764f437a7835.pem
rm /tmp/b.pem*; ssh-keygen -f /tmp/b.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/b.pem.pub -e -m pem > /tmp/authenticators/ed94a517-f030-4aa6-b390-6283ad24cab3.pem
rm /tmp/ticker.pem*; ssh-keygen -f /tmp/ticker.pem -t rsa -m pem -N "" && ssh-keygen -f /tmp/ticker.pem.pub -e -m pem > /tmp/authenticators/ticker.pem

aws dynamodb list-tables --endpoint-url http://localhost:8000 --region us-west-2

if [[ -n ${DEBUG_TICKER} ]]; then
    read -p "Start debugger for ticker, then press any key to continue: " NOTHING
else
    ${ROOTDIR}/bin/ticker  --grpcPort 8001 --httpPort 8002 -kvConnectionString "dynamo:endpoint=http://localhost:8000,region=us-west-2,tableName=tickerKVStore,prefixLength=4" \
        -authenticatorPath /tmp/authenticators -privateKeyPath /tmp/ticker.pem > /tmp/ticker.out 2>&1 &
    TICKER_PID=$!
fi

sleep 2

curl --silent -X POST http://localhost:8002/api/v1/tick > /dev/null 2>&1

if [[ -n ${DEBUG_BINGE} ]]; then
    read -p "Start debugger for binge, then press any key to continue: " NOTHING
else
    ${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort 8008 -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${ROOTDIR}/test/functional/entwine/config/binge-a.json > /tmp/binge-a.out 2>&1 &
    BINGE1_PID=$!
fi

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort 8009 -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${ROOTDIR}/test/functional/entwine/config/binge-b.json > /tmp/binge-b.out 2>&1 &
BINGE2_PID=$!
sleep 2


let NUM_MESSAGES=100

let i=0
let A_index=1
let B_index=1
declare -a INDEXES
while (( ${i} < ${NUM_MESSAGES} )); do
    choose=$(($RANDOM % 2))
    if (( ${choose} == 0 )); then
        payload='{"idx":'${i}', "aIdx": '${A_index}'}'
        echo ${payload} | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8008/entwine -o /dev/null
        INDEXES[${i}]=${A_index}
        let A_index=${A_index}+1
    else
        payload='{"idx":'${i}', "bIdx": '${B_index}'}'
        echo ${payload} | curl --silent -H "Content-Type: application/json" -X POST --data-binary @- http://localhost:8009/entwine -o /dev/null
        INDEXES[${i}]=${B_index}
        let B_index=${B_index}+1
    fi
    let i=${i}+1
done 

declare -a MESSAGES
for outfile in `echo /tmp/a-out/* /tmp/b-out/*`; do
    let idx=$(cat ${outfile} | jq -r '.idx')
    uuid=`cat ${outfile} | jq -r '.["internal:entwine:output"][0].entwineUuid'`
    subStreamID=`cat ${outfile} | jq -r '.["internal:entwine:output"][0].subStreamID'`
    MESSAGES[${idx}]="${subStreamID}:${uuid}:"${INDEXES[${idx}]}
done
    
let num_success=0
let i=${i}-1
while (( ${i} > 0 )); do
    result=`${ROOTDIR}/bin/entwinectl  -tickerEndpoint localhost:8001 -command HappenedBefore ${MESSAGES[$((i))]} ${MESSAGES[$((i-1))]} | egrep -v -E '[0-9a-z]{8}\-[0-9a-z]{4}\-[0-9a-z]{4}\-[0-9a-z]{4}\-[0-9a-z]{12} happenedBefore [0-9a-z]{8}\-[0-9a-z]{4}\-[0-9a-z]{4}\-[0-9a-z]{4}\-[0-9a-z]{12}: false'`
    if [[ -z ${result} ]]; then
        let num_success=${num_success}+1
    else
        echo "Error evaluating happened before relationship: ${result}"
    fi
    let i=${i}-1
done

ret=0
if (( ${num_success} == ${NUM_MESSAGES} - 1 )); then
    echo "RESULT: success"
else
    echo "RESULT: failure"
    ret=1
fi

kill -9 ${TICKER_PID} ${BINGE1_PID} ${BINGE2_PID}

exit ${ret}
