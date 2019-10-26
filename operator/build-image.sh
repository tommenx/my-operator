#!/usr/bin/env bash

docker build ./ -t registry:5000/tidb-operator:latest
docker push registry:5000/tidb-operator:latest