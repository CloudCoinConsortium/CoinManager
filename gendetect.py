#!/usr/bin/python3



import requests
import json
import time


#data={"coins":[{"sn":155, "ans":["10000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000", "00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000","00000000000000000000000000000000" ]}]}

data={
"coins":[]
}

ncoins=1062
for i in range(0, ncoins):
	sn=154+i
	k={
		"sn":sn,
		"ans":[None]*25,
	}
	for r in range(0, 25):
		k['ans'][r]="00000000000000000000000000000000"

	data['coins'].append(k)


print(data)

strd=json.dumps(data)

print(strd)

url="http://localhost:8888/api/v1/detect"

headers={
"Content-Type":"application/json"
}

r = requests.post(url=url, data=strd, headers=headers)

o=json.loads(r.text)
taskid=o['payload']['id']

print("r="+taskid)


while True:
	url="http://localhost:8888/api/v1/tasks/" + taskid
	r = requests.get(url=url)
	print(r.text)
	time.sleep(1)
