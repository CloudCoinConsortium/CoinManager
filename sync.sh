#!/bin/bash 




#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"ddd", "amount":524, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":2, "tag":"xxx", "toname":"miroch6.skyvault.cc"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":151, "tag":"xxx", "to":2}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/sync' -d '{"name":"Default", "sync_items" : {"2301588":[1,1,1,1,1,2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1]}}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/sync' -d '{"name":"miroch23.skyvault.cc", "sync_items": {"2301588":[1,1,1,1,1,2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1],  "123":[1,2,3,4,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1,1,1,1,1], "12345":[2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,2,1,2,2,2,2]}  }' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/deposit' -d '{"name":"ddd", "amount":100, "tag":"xxx", "to":2}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/sync' -d '{"name":"miroch30.skyvault.cc", "sync_items": {"2301714":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,3,1,1,1,1,1,1,1,1,2,1],"2301715":[1,1,1,1,1,1,1,1,1,1,1,1,1,1,3,1,1,1,1,1,1,1,1,2,1]}}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/sync' -d '{"name":"miroch87.skyvault.cc", "sync_items": {"230009":[2,1,1,1,1,1,1,1,1,1,1,1,1,1,3,1,1,1,1,1,1,1,1,5,1]}}' 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/sync' -d '{"name":"miroch87.skyvault.cc", "sync_items": { "230010":[2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1,1,5,5,1],"230011":[2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1,1,5,5,1],"230012":[2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1,1,5,5,1],"230013":[2,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,1,2,1,1,5,5,1]   }}' 2>/dev/null| jq -r '.payload.id')



while [ 1 ]; do
  r=$(curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null)

  echo $r
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





