
## FABUSERS ##

description: TODO

## PREREQUISITES ##

You have to install prerequisites as it is described [here](https://hyperledger-fabric.readthedocs.io/en/latest/prereqs.html)
and [platform-specific binaries](https://hyperledger-fabric.readthedocs.io/en/latest/samples.html#binaries).

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

3. According to [tutorial](https://hyperledger-fabric.readthedocs.io/en/latest/write_first_app.html),
an *admin* user was registered with CA. But now we need retrieve the eCert for the *admin*.

		node enrollAdmin.js

4. Register new user (*user1*) in the system

		node registerUser.js

5. Invoke request needs to change *invoke.js*, and:

		node invoke.js

6. For query, you should change *query.js*. After that:

		node query.js


