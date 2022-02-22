#!/usr/bin/python


import sys

total=1000

start=int(sys.argv[1])

for sn in range(start, start+total):
    fname="/home/alexander/dev/superraida/coins/" + str(sn) + ".bin"


    f = [0, 0, 0, 1, 0, 0]
    for z in range(0, 10):
        f.append(0)

    for z in range(0, 16):
        f.append(0)

    f.append((sn >> 16) &0xff)
    f.append((sn >> 8) &0xff)
    f.append(sn & 0xff)

    for r in range(0, 13):
        f.append(0xaa)

    for r in range(0, 25):
        for x in range(0, 16):
            f.append(0)

    f = bytearray(f)

    nf = open(fname, "wb")
    nf.write(f)
    nf.close()




print start
