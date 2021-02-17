 docker run -d -p 8000:8000 amazon/dynamodb-local

aws dynamodb create-table --endpoint-url http://localhost:8000 --region us-west-2 \
     --table-name localkvstore \
     --attribute-definitions \
         AttributeName=KeyPrefix,AttributeType=S \
         AttributeName=ValueKey,AttributeType=S \
     --key-schema AttributeName=KeyPrefix,KeyType=HASH AttributeName=ValueKey,KeyType=RANGE \
     --provisioned-throughput ReadCapacityUnits=1,WriteCapacityUnits=1
