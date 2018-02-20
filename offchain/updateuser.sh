#!/bin/bash

curl -X PUT -H "Content-Type: application/json" -d @newuserinfo.json http://localhost:8080/users/ondar07?password=superPassword