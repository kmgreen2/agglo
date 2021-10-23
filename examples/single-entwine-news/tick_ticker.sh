while true
do
  curl --silent -X POST http://localhost:8002/api/v1/tick > /dev/null 2>&1
	sleep 2
done