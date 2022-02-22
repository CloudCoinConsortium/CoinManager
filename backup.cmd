#!/bin/bash



TASKID=$(curl 'http://localhost:8888/api/v1/backup'  -d '{"name":"Default", "folder":"/home/alexander/ccbackup"}' | jq -r '.payload.id')


echo $TASKID
while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
    sleep 1
    done

