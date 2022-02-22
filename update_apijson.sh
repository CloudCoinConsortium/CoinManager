#!/bin/sh


sudo sh -c 'mv /media/sf_common/SuperRAIDA-Client-API* api.json'
sudo sh -c 'mv /media/sf_common/cloudcoin_manager*.msi cloudcoin_manager.msi'
sudo chmod 644 api.json cloudcoin_manager.msi
sudo chown alexander:alexander api.json cloudcoin_manager.msi

