#!/bin/bash

#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":1234, "pownstring":"ppppppppppppppppppppppppp" }]}')
#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":11199, "pownstring":"pppppppppppppppppppppppfp" }]}' |  jq -r '.payload.id'   )
#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":11199, "pownstring":"pppppppppppppppppppppppfp" }]}' |  jq -r '.payload.id'   )


TASKID=$(curl 'http://localhost:8888/api/v1/skyfix' -d '{"name":"miroch77.skyvault.cc", "pownstring":"fpfpppppppppupppppppppupn"}' | jq -r '.payload.id')

echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done
