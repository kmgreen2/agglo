ROOTDIR=$( dirname $0 )/../..
THISDIR=$( dirname $0 )

MINIO_PORT=9000

export AWS_ACCESS_KEY_ID=localtest
export AWS_SECRET_ACCESS_KEY=localtest
export AWS_DEFAULT_REGION=us-west-2

docker build . -t tokenizer

../../bin/dumbserver -port 1337 >/tmp/dumbserver.1337 2>&1 &
PID1=$!

docker run -d -p ${MINIO_PORT}:${MINIO_PORT} \
    -e "MINIO_ROOT_USER=localtest" \
    -e "MINIO_ROOT_PASSWORD=localtest" \
    --name minio \
    minio/minio server /data || exit 1

sleep 2
mc alias set localtest http://localhost:9000 localtest localtest

# Create buckets to use in the test
mc mb localtest/localtest || exit 1

../../bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 4 -runType stateless-daemon -exporter none -config ./nlp-pipeline.json >/tmp/binge.out 2>&1 &
PID2=$!

sleep 5

../../bin/genevents  -numEvents 20 -numThreads 1 -schema ./nlp-pipeline-gen.json -output http://localhost:80/binge

kill -9 ${PID1} ${PID2}
docker kill minio
docker rm minio
