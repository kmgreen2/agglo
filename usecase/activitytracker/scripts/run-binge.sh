ROOTDIR=$( dirname $0 )/../../..

${ROOTDIR}/bin/binge  -daemonPath /webhook -daemonPort 80 -maxConnections 16 -runType persistent-daemon -exporter zipkin -config ${ROOTDIR}/usecase/activitytracker/config/activitytracker.json
