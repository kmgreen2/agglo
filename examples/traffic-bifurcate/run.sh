../../bin/dumbserver -port 1337 >/tmp/dumbserver.1337 2>&1 &
PID1=$!
../../bin/dumbserver -port 1338 >/tmp/dumbserver.1338 2>&1 &
PID2=$!
../../bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 4 -runType stateless-daemon -exporter none -config ./traffic-bifurcate.json >/tmp/binge.out 2>&1 &
PID3=$!

sleep 5

../../bin/genevents  -numEvents 20 -numThreads 1 -schema ./traffic-bifurcate-gen.json -output http://localhost:80/binge

kill -9 ${PID1} ${PID2} ${PID3}

