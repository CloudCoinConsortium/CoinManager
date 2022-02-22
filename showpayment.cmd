#!/bin/bash



#TASKID=$(curl 'http://localhost:8888/api/v1/skywallets/miroch8.skyvault.cc'  | jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/showpayment' -d '{"guid":"b09ec4caeed8678f34243e3a5155acea"}' | jq -r '.payload.id')


echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
    sleep 1
    done

