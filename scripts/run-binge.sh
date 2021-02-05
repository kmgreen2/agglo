# GOTRACEBACK=crash bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 16 -runType persistent-daemon -exporter zipkin -config test/config/cicd_pipeline.json  
 GOTRACEBACK=crash bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 16 -runType stateless-daemon -exporter zipkin -config test/config/cicd_pipeline.json  
# GOTRACEBACK=crash bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 16 -numThreads 32 -runType persistent-daemon -exporter zipkin -config test/config/cicd_pipeline.json 
# GOTRACEBACK=crash bin/binge  -daemonPath /binge -daemonPort 80 -maxConnections 16 -numThreads 16 -runType stateless-daemon -exporter zipkin -config test/config/slow_spawner_pipeline.json
