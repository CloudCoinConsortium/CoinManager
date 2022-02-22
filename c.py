#!/usr/bin/python

import re

with open('./skywallet.cc.zone') as f:
    lines = f.readlines()


for l in lines:
    l = l.rstrip('\n')
        
    l = re.split('\s+', l)
    if (len(l) != 3):
        continue

    if (l[1] != 'A'):
        continue

    ip = l[2]

    o = ip.split('.')

    b1 = int(o[1])
    b2 = int(o[2])
    b3 = int(o[3])
    #sn = (o[1]<<16)|(o[2]<<8)|int(o[3])
    sn = (b1<<16)|(b2<<8)|b3
    print sn

