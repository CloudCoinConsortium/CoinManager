#!/bin/bash 




b="AAAAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAABDmN3SubxKUDprNwgpQgRCToxhBD7r5A8UV8M7JZ4O5/usYPsNwbzklb7jhsQ/xV3Jv83yffj5aKWJF0b7xd1b4Kjcd0KjFpH0rYUyCPsziGekNirP4YpQgjO2r/wIktscU0OubUABvnMQSMTQ/ZUElk+YbPweixC05P53lvM53ebM+mXe5G+VaJ3Bcs6Ac70TfCMQpYowtxGAJNLasuFL6QULNAQRBTfjfZxud2oQ5oA3SU4BsvMzVvLF+A5cr6waBufVWTJ1rkVSOs0oJTCZZoiSJnO4cXOIO8gjh2P+kQhisje6M7GLYuaTyztMzMY2IFuquYyDyrfIAHXLnv0Q/SXRQ4IVdK3nmCCyR5kOtqyW+S7xv/LsufAYu2ED16ipij2KSkqP/Iv8OIACAk2TsRvTVDudGHHUt2a8z4l47ynIaadC/duM0uhHgr4JkLB9ZLy1N0EXpQW1x10gJdZt9mRfmoBHTv6gV/9N2fSuqdlSL3FJjHbzLnAOnVBX6GBo/Yn24LVDL4gOcglDLT2X"

#b=$(cat /home/alexander/axx.skywallet.cc.png |base64 -w0)

#curl -v 'http://localhost:8888/api/v1/unpack' -d "{\"data\":\"$b\"}"
#TASKID=$(curl 'http://localhost:8888/api/v1/import' -d @<(echo "{\"name\":\"ddd\", \"items\":[{\"type\":\"inline\", \"data\":\"$b\"} ]}") 2>/dev/null| jq -r '.payload.id')
TASKID=$(curl 'http://localhost:8888/api/v1/import' -d @<(echo "{\"name\":\"ddd\", \"items\":[{\"type\":\"file\", \"data\":\"/home/alexander/dev/superraida/file.zip\"} ]}") 2>/dev/null| jq -r '.payload.id')


while [ 1 ]; do
  curl "http://localhost:8888/api/v1/tasks/$TASKID" 2>/dev/null
  sleep 1
done





