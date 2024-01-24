#!/bin/sh

REG=lccro/tv-gin:latest
docker build --platform linux/amd64 -t $REG .
docker push $REG

