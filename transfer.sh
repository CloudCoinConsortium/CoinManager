#!/bin/bash 




#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"ddd", "amount":524, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/transfer' -d '{"srcname":"ddd331", "dstname":"test", "amount":2, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')

echo $TASKID
while [ 1 ]; do
  r=$(curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null)

  status=$(echo $r | jq -r '.payload.status')

  echo "st=$status"
  if [ "$status" = "completed" ]; then
    echo $r

    echo "done"

    exit

  fi

  sleep 1
done





