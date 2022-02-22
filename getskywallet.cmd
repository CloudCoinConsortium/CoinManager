#!/bin/bash



#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch11.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch13.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch15.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch16.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch31.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch43.skyvault.cc'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/amiroch42.skyvault.cc'  | jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch87.skyvault.cc'  | jq -r '.payload.id')


echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/${TASKID}"
    sleep 1
    done

