docker run -d -p 9000:9000 \
  -e "MINIO_ROOT_USER=localtest" \
  -e "MINIO_ROOT_PASSWORD=localtest" \
  --name minio \
  minio/minio server /data

mc alias set localtest http://localhost:9000 localtest localtest

# Create buckets to use in the test
mc mb localtest/localtestA
mc mb localtest/localtestB

