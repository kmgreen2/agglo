GOTRACEBACK=crash bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 16 -runType stateless-daemon -exporter zipkin -config test/config/cicd_pipeline.json  
