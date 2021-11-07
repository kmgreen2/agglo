#!/bin/bash

ROOTDIR=$( dirname $0 )/../..

docker kill dynamodb
docker rm dynamodb
docker kill minio
docker rm minio

${ROOTDIR}/deployments/local/elastic/stop-elastic.sh

rm /tmp/a-out/*
rm /tmp/b-out/*

rm /tmp/authenticators/*
rmdir /tmp/authenticators

rm /tmp/a.pem*
rm /tmp/b.pem*

rm /tmp/ticker.out
