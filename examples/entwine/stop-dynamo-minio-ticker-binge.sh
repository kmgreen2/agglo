#!/bin/bash

docker kill dynamodb
docker rm dynamodb
docker kill minio
docker rm minio

rm /tmp/a-out/*
rm /tmp/b-out/*

rm /tmp/authenticators/*
rmdir /tmp/authenticators

rm /tmp/a.pem*
rm /tmp/b.pem*

rm /tmp/binge-a.out
rm /tmp/binge-b.out
rm /tmp/ticker.out
