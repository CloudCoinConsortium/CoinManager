#!/bin/bash 




#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"ddd", "amount":524, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"ddd", "amount":2, "tag":"xxx"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"Default", "amount":2, "tag":"xxx", "folder":"/home/alexander/ttt"}' 2>/dev/null| jq -r '.payload.id')
#TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"errwe", "amount":2, "tag":"yyy", "type":"bin", "folder":"/home/alexander/ttt"}' 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/export' -d '{"name":"Default", "amount":1, "tag":"yyy", "type":"png", "folder":"/home/alexander/ttt"}' 2>/dev/null| jq -r '.payload.id')


while [ 1 ]; do
  r=$(curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null)

  status=$(echo $r | jq -r '.payload.status')

  echo "st=$status"
  if [ "$status" = "completed" ]; then
    echo $r | jq -r '.payload.data.coins' |base64 -d > ./file.png



    echo "done"

    exit

  fi

  sleep 1
done





