ROOTDIR=$( dirname $0 )/../..
THISDIR=$( dirname $0 )

function cleanup {
  kill -9 ${TICKER_PID}
  rm ${THISDIR}/us-test-corpus-2017-08-17-2wks.csv
}

trap cleanup EXIT INT TERM

unzip ${THISDIR}/us-test-corpus-2017-08-17-2wks.zip

${THISDIR}/tick_ticker.sh &
TICKER_PID=$!

${ROOTDIR}/bin/csv2json -concurrency 8 -outEndpoint http://localhost:8008/entwine -csvFile ${THISDIR}/us-test-corpus-2017-08-17-2wks.csv -csvMap 0:created,1:source,2:url,3:text

