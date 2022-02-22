#!/bin/bash
TASKID=$(curl -v 'http://localhost:8888/api/v1/skydetect' -d '{"name":"miroch77.skyvault.cc"}' | jq -r '.payload.id')



while [ 1 ]; do
  curl -v "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done
