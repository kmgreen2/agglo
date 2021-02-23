ROOTDIR=$( dirname $0 )/../../..

${ROOTDIR}/bin/ticker  --grpcPort 8001 -kvConnectionString "dynamo:endpoint:localhost:8000,region=us-west-2,tableName=tickerKVStore,prefixLength=4" \
    -authenticatorPath /tmp/authenticators -privateKeyPath /tmp/ticker.pem
