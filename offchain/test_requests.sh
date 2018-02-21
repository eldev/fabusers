#!/bin/bash


# 1. To add a new user with data described in userinfo.json:
curl -X POST -H "Content-Type: application/json" -d @userinfo.json http://localhost:8080/users

# 2. To receive user data
#    Moreover, if password matches user password or admin password, this request get private user data (decrypted data)
curl -H "Content-Type: application/json" http://localhost:8080/users/ondar07?password=superPassword

# 3. TO update user data with newuserinfo.json
#    If password doesn't matches to admin's password the service will not update data
curl -X PUT -H "Content-Type: application/json" -d @newuserinfo.json http://localhost:8080/users/ondar07?password=AdminSuperPassword