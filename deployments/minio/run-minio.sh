docker run -d -p 9000:9000 \
  -e "MINIO_ROOT_USER=localtest" \
  -e "MINIO_ROOT_PASSWORD=localtest" \
  minio/minio server /data

mc alias set localtest http://localhost:9000 localtest localtest

mc mb localtest localtest
