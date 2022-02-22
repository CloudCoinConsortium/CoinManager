#!/bin/bash



TASKID=$(curl 'http://localhost:8888/api/v1/withdraw'  -d '{"srcname":"miroch40.skyvault.cc", "dstname":"Default", "amount":1, "tag":"a1111"}' | jq -r '.payload.id')


echo $TASKID
while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
    sleep 1
    done

