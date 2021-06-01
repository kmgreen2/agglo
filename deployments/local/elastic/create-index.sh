BASEDIR=`dirname $0`

curl -X PUT "http://localhost:9200/test-index" -H 'Content-Type: application/json' -T ${BASEDIR}/schema.json 
