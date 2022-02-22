#!/bin/bash
#TASKID=$(curl -v 'http://localhost:8888/api/v1/detect' -d '{"coins":[{"sn":155, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]},  {"sn":156, "ans":["00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]},       {"sn":157, "ans":["00000000000000000000000000000001", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]}   , {"sn":158, "ans":["00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]} ,             {"sn":159, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]} , {"sn":160, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]},   {"sn":161, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]} ,   {"sn":162, "ans":["00000000000000000000000000000001", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]},  {"sn":163, "ans":["00000000000000000000000000000000",   ]}'  | jq -r '.payload.id')
#TASKID=$(curl -v 'http://localhost:8888/api/v1/detect' -d '{"coins":[{"sn":11155, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ], "coin_type":0}   ]}'  | jq -r '.payload.id')





TASKID=$(curl -v 'http://localhost:8888/api/v1/detect' -d '{"coins":[{"sn":53137, "ans":[        "a70c24e19386aad38f9bc1e1d47c34e5",         "502871dc8e54ffde1ca066b40d237f31",        "4bdb57d7508af6b4a7db623268e4ee10",        "bd451d1a28b89c49bfdd19e97928d23e",        "0cfe2c71b5f97a52b44ed184d6559b27",        "e34ffb2defadd7bf7d5e727fbd72cd2d",        "28924dc3dcda3f5fcd8d019d6ede420d",        "7e01346202116b4e13d2e024b0ead091",        "e835b5114e920e5df98ed196a62e6867",        "1eec8268d0e1683396d9dcee7abd3920",        "37093ab1936bd362ea3ba244f64d2b76",        "3c15546d2159513d7a17059416a6e748",        "fc2c9ac8374b14f5b290924a460e8075",        "41a526669f9b577eb7cd3065f5c7ccab",        "f8acea0ff8b11e16e06933a7825389a6",        "a27f5e51966462b46c20c02bdb7c7f52",        "6af78ee816fad2acf8f8082854fd38bc",        "c40696400164bb6b06fab83514714a2d",        "e0facfc725e00b71ff30e52a4f24adbd",        "874bedec832489c2e1dd4d935c79e7de",        "fe2ab8bd138dce58e6a048a2ea7d50d9",        "2eee4b76fe171d8015e7665577b99b9e",        "3936e4c78c8d0d6ff3450cdd5f3a202f",        "6bcf5a9d6283eaea6a6495e16b4da957",        "95c8a27cad9e2f04197ab01878770afb"], "coin_type":0}   ]}'  | jq -r '.payload.id')



while [ 1 ]; do
  curl -v "http://localhost:8888/api/v1/tasks/$TASKID"
  sleep 1
done
