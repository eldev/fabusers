
## FABUSERS ##

The Fabusers project stores user data (public and private data).
It consists of 2 main parts:
1. blockchain part (using [Hyperledger Fabric](https://www.hyperledger.org/projects/fabric)) contains records of the key-value type (where the key is a user name, a value is a hash value of his offchain data).
2. offchain part uses mongodb to store data.


## PREREQUISITES ##

You have to install prerequisites as it is described [here](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html)
and [platform-specific binaries](https://hyperledger-fabric.readthedocs.io/en/latest/samples.html#binaries).

The offchain part uses a mongodb database so you have to install (mgo package (MongoDB driver for Golang)):
		go get gopkg.in/mgo.v2
		go get gopkg.in/mgo.v2/bson

In addition, goji package:
		go get goji.io

After that, you should copy the project's chaincode sample into $GOPATH directory:
		cp ./fabusers_chaincode/fabusers.go $GOPATH/src/fabusers/fabusers.go

## HOW TO RUN? ##

1. Launch the network (Fabric entities and chaincode container)
This script also install and instantiate the chaincode
(This command may require root permission)

		./fabusers/startFabric.sh

2. Open *./fabusers* directory

		cd ./fabusers/
	
3. Install the necessary packages

		npm install

4. Open ./offchain, and build this part:

		go build fabusers_srv.go

5. Run the main service process in terminal 1:

		./fabusers_srv

6. In terminal 2 you can send requests to the service by using curl utility and json files.
For example, to add user with data described in userinfo.json:

		curl -X POST -H "Content-Type: application/json" -d @userinfo.json http://localhost:8080/users

Other examples of requests you can see in *test_requests.sh*.


