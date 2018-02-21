'use strict';

/*
 * Chaincode query
 *
 * USER:
 *     node query.js <userlogin>
 */

var Fabric_Client = require('fabric-client');
var path = require('path');
var util = require('util');
var os = require('os');

//
var fabric_client = new Fabric_Client();

// setup the fabric network
var channel = fabric_client.newChannel('mychannel');
var peer = fabric_client.newPeer('grpc://localhost:7051');
channel.addPeer(peer);

//
var member_user = null;
var store_path = path.join(__dirname, 'hfc-key-store');
var tx_id = null;

// query for this user
var userlogin = process.argv[2]

// create the key value store as defined in the fabric-client/config/default.json 'key-value-store' setting
Fabric_Client.newDefaultKeyValueStore({ path: store_path
}).then((state_store) => {
	// assign the store to the fabric client
	fabric_client.setStateStore(state_store);
	var crypto_suite = Fabric_Client.newCryptoSuite();
	// use the same location for the state store (where the users' certificate are kept)
	// and the crypto store (where the users' keys are kept)
	var crypto_store = Fabric_Client.newCryptoKeyStore({path: store_path});
	crypto_suite.setCryptoKeyStore(crypto_store);
	fabric_client.setCryptoSuite(crypto_suite);

	// get the enrolled user from persistence, this user will sign all requests
	return fabric_client.getUserContext(userlogin, true);
}).then((user_from_store) => {
	if (user_from_store && user_from_store.isEnrolled()) {
		// Successfully loaded @userlogin from persistence
		member_user = user_from_store;
	} else {
		throw new Error('Failed to get userlogin.... run registerUser.js');
	}

	// queryUser chaincode function - requires 1 argument, ex: args: ['user1'],
	const request = {
		//targets : --- letting this default to the peers assigned to the channel
		chaincodeId: 'fabusers',
		fcn: 'queryUser',
		args: [userlogin]
	};

	// send the query proposal to the peer
	return channel.queryByChaincode(request);
}).then((query_responses) => {
	// Query has completed, checking results
	// query_responses could have more than one  results if there multiple peers were used as targets
	if (query_responses && query_responses.length == 1) {
		if (query_responses[0] instanceof Error) {
			console.error("error from query = ", query_responses[0]);
		} else {
			console.log("OK RESPONSE:", query_responses[0].toString());
		}
	} else {
		console.log("No payloads were returned from query");
	}
}).catch((err) => {
	console.error('Failed to query :: ' + err);
});
