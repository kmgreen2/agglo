THISDIR=$(dirname $0)

if [[ -n `docker ps | egrep 'broker|zookeeper'` ]]; then
    echo "Detected running broker and/or zookeeper, please kill those processes..."
    exit 1
fi

docker-compose -f ${THISDIR}/docker-compose.yml up -d
