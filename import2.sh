#!/bin/bash 


x='{
        "cloudcoin": [{
                "nn":"0",
                "sn":"4301956",
                "an":["00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","02000000000000000000000000000000","00000000000000000000000000000000"],
                "pown": "ppppppppppppppppppppupppp",
                "aoid": []
        }, {
                "nn":"1",
                "sn":"4301957",
                "an":["00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000"],
                "pown": "ppppppppppppppppppppupppp",
                "aoid": []
        }
        ]
}'


b=$(echo $x |base64 -w0)


#b=$(cat /home/alexander/axx.skywallet.cc.png |base64 -w0)

#curl -v 'http://localhost:8888/api/v1/unpack' -d "{\"data\":\"$b\"}"
TASKID=$(curl 'http://localhost:8888/api/v1/import' -d @<(echo "{\"name\":\"ddd331\", \"items\":[{\"type\":\"inline\", \"data\":\"$b\"} ]}") 2>/dev/null| jq -r '.payload.id')


while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null
  sleep 1
done





