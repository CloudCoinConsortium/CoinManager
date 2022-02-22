#!/bin/bash

TASKID=$(curl 'http://localhost:8888/api/v1/echo' | jq -r '.payload.id')



while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done

