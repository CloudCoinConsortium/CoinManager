#!/usr/bin/python3

import os, sys


if len(sys.argv) != 2:
    print("provide start sn")
    sys.exit(1)

start=int(sys.argv[1])
ncoins=3000
dir="/home/alexander/dev/superraida/coins"

def genCoin(sn):

    b = bytearray()
    b.append(0x0)
    b.append(0x0)

    # CoindID
    b.append(0x0)
    b.append(0x1)

    # Split ID
    b.append(0x0)

    # Enc ID
    b.append(0x0)

    # PasswordHash
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)

    # Flags
    b.append(0x0)

    # ReceiptID
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)
    b.append(0x0)

    # Data
    b.append((sn >> 16) & 0xFF)
    b.append((sn >> 8) & 0xFF)
    b.append(sn & 0xFF)

    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)
    b.append(0xAA)

    for r in range(0,25):
        for an in range(0,16):
            b.append(0x0)


    fname = dir + "/" + str(sn) + ".bin"
    f = open(fname, "wb")
    f.write(b)
    f.close()

for sn in range(start, start + ncoins):
    genCoin(sn)


