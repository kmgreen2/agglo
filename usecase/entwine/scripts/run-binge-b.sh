ROOTDIR=$( dirname $0 )/../../..

${ROOTDIR}/bin/binge  -daemonPath /entwine -daemonPort 80 -maxConnections 16 -runType stateless-daemon -exporter stdout -config ${ROOTDIR}/usecase/binge/config/binge-b.json
