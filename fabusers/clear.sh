#!/bin/bash

# kill any stale or active containers
docker rm -f $(docker ps -aq)

# clear any cached networks
docker network prune

docker volume prune

rm -rf node_modules

rm -rf hfc-key-store

rm -rf ../chaincode

docker image rm dev-peer0.org1.example.com-fabusers-1.0-aec82f59b5c2011821cc8038aec29bfb5bf138d872c7b7f91cd0aa90b5f37344
