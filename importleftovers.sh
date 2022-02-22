#!/bin/bash 



#TASKID=$(curl 'http://localhost:8888/api/v1/import' -d @<(echo "{\"name\":\"Default\", \"items\":[{\"type\":\"suspect\"}]}") 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/import' -d @<(echo "{\"name\":\"ddd331\", \"items\":[{\"type\":\"suspect\"}]}") 2>/dev/null| jq -r '.payload.id')


while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null
  sleep 1
done





