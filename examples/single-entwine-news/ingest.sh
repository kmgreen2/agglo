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

subStreams="e546e731-2b86-43cd-b847-764f437a7835 e546e731-2b86-43cd-b847-764f437a7834 e546e731-2b86-43cd-b847-764f437a7833 e546e731-2b86-43cd-b847-764f437a7832 e546e731-2b86-43cd-b847-764f437a7831 e546e731-2b86-43cd-b847-764f437a7830 e546e731-2b86-43cd-b847-764f437a7829"

sourceMatch="w3-businessinsider w3-foxnews w3-huffingtonpost w3-latimes w3-new-yahoo w3-reuters w3-siouxcityjournal"
currDate="2017-08-24"
endDateExclusive="2017-09-01"

while [[ ${currDate} != ${endDateExclusive} ]]; do
    ${ROOTDIR}/bin/csv2json -concurrency 1 -outEndpoint http://localhost:8008/entwine -csvFile ${THISDIR}/us-test-corpus-2017-08-17-2wks.csv -csvMap 0:created,1:source,2:url,3:text &

    currDate=`date  -j -f %Y-%m-%d -v+1d ${currDate} +'%Y-%m-%d'`
done



