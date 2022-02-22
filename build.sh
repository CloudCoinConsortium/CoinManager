#!/bin/bash 


MDIR=/home/alexander/dev/superraida
DIR=/home/alexander/dev/srui/cloudcoin-ng/dist\ folder/cloud-wallet
TDIR=/media/sf_common/cloudcoin_manager


cd /home/alexander/dev/srui/cloudcoin-ng
git pull origin anjali

list=$(sudo sh -c "echo $TDIR/*")
for i in $list; do
  sudo rm -rf $i
done

sudo cp -a "$DIR"/* $TDIR

sudo cp -a $MDIR/backassets $TDIR
sudo cp -a $MDIR/cloudcoin_manager.{exe,bat} $TDIR
sudo cp -a $MDIR/rdll/*.dll $TDIR
