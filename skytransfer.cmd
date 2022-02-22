#!/bin/bash



#TASKID=$(curl 'http://localhost:8888/api/v1/skytransfer' -d '{"srcname":"xxxxxxxxxxx", "dstname":"axx.skywallet.cc", "amount":10, "tag":"dfgdsg"}'  | jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/skytransfer' -d '{"srcname":"miroch30.skyvault.cc", "dstname":"miroch31.skyvault.cc", "amount":5, "tag":"takeit2"}'  | jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/skytransfer' -d '{"srcname":"miroch71.skyvault.cc", "dstname":"miroch70.skyvault.cc", "amount":2, "tag":"takeit2"}'  | jq -r '.payload.id')



echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
    sleep 1
    done

