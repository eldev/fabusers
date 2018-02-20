#!/bin/bash

# This script 

# kill any stale or active containers
docker rm -f $(docker ps -aq)

# clear any cached networks
docker network prune

docker volume prune

rm -rf node_modules

rm -rf hfc-key-store

rm -rf ../chaincode
