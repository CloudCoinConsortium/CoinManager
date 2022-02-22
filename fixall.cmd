#!/bin/bash

#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":1234, "pownstring":"ppppppppppppppppppppppppp" }]}')
#TASKID=$(curl -X PUT 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd331"}' |  jq -r '.payload.id'   )
#TASKID=$(curl -X PUT 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd331"}' |  jq -r '.payload.id'   )
TASKID=$(curl -X PUT 'http://localhost:8888/api/v1/fix' -d '{"name":"Default"}' |  jq -r '.payload.id'   )

echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done
