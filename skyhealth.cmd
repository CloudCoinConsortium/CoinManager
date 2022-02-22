#!/bin/bash



#TASKID=$(curl 'http://localhost:8888/api/v1/skyhealth'  -d '{"name":"miroch77.skyvault.cc"}' | jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/skyhealth'  -d '{"name":"miroch87.skyvault.cc"}' | jq -r '.payload.id')


echo $TASKID
while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
    sleep 1
    done

