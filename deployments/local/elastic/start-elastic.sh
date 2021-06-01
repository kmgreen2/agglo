BASEDIR=`dirname $0`

docker-compose -f ${BASEDIR}/docker-compose.yml up --force-recreate -d

${BASEDIR}/create-index.sh
is_elastic_up=$?
while (( ${is_elastic_up} != 0 )); do
    sleep 5
    ${BASEDIR}/create-index.sh
    is_elastic_up=$?
done

