#!/bin/bash

#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":1234, "pownstring":"ppppppppppppppppppppppppp" }]}')
#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":11199, "pownstring":"pppppppppppppppppppppppfp" }]}' |  jq -r '.payload.id'   )
#TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"ddd", "pown_items":[{"sn":11199, "pownstring":"pppppppppppppppppppppppfp" }]}' |  jq -r '.payload.id'   )


TASKID=$(curl 'http://localhost:8888/api/v1/fix' -d '{"name":"Default","pown_items":[{"sn":1001,"pownstring":"fppppppppppfupppppppppupn"}],"tickets":[["","53989a8a","5409b025","23c737d1","51fd2b29","034c3c3b","258c667e","64ac93eb","586edf53","4a1c8b1c","6f5718d8","","","3f31cd35","49d60e44","60ebe6fb","61de684b","349ce890","2a57c540","1585e95b","2e7015a7","56afa325","","7574d8d5",""]]}' | jq -r '.payload.id')

echo $TASKID

while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done
