#!/bin/bash 




#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"ddd", "amount":524, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":2, "tag":"xxx", "toname":"miroch6.skyvault.cc"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":151, "tag":"xxx", "to":2}' 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd331", "amount":1, "tag":"take it", "toname":"miroch39.skyvault.cc"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":100, "tag":"xxx", "to":2}' 2>/dev/null| jq -r '.payload.id')


while [ 1 ]; do
  r=$(curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null)

  status=$(echo $r | jq -r '.payload.status')

  echo "st=$status"
  if [ "$status" = "completed" ]; then
    #echo $r | jq -r '.payload.data.coins' |base64 -d > ./file.png
    echo $r | jq 



    echo "done"

    exit

  fi

  sleep 1
done





