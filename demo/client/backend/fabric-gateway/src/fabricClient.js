/*
* Copyright IBM Corp All Rights Reserved
*
* SPDX-License-Identifier: Apache-2.0
*/

'use strict';

const FabricCAClient = require('fabric-ca-client');
const { Gateway, FileSystemWallet, X509WalletMixin } = require('fabric-network');
const fs = require('fs');
const path = require('path');


/////////////////  Global constants  /////////////////
//  errorcodes
const SUCCESS = 0;
const USER_UNAUTHORIZED = 401;
const USER_EXISTS = 402;
const FAILURE  = 499;

//  global variables loaded initially
let configdata;
let connectionProfile;
let channelName;
let chaincodeID;
let wallet;
let adminUserName;
let secret;
let bLocalHost;

//  other global variables
let adminGateway;
let contract;

function prepareStatus (returncode, message) {
    let result =  { 'status': {  'rc': returncode, 'msg': message } };
    return result;
}

function getRole (identity) {
    let attrs = (identity.attrs).filter ( function (attr) {
        return attr.name === 'approle';
    }) ;
    if ( (attrs) && (attrs.length > 0) )
    { return attrs[0].value; }
    else { return '';             }
}

//  get client's MSPID from connectionProfile
function getClientMspid () {
    let clientOrgName = connectionProfile.client.organization;
    return connectionProfile.organizations[clientOrgName].mspid;
}

//  get client's organization's CA URL from connectionProfile
function getClientCAUrl () {
    let clientOrgName = connectionProfile.client.organization;
    let ca = connectionProfile.organizations[clientOrgName]
        .certificateAuthorities[0];
    return connectionProfile.certificateAuthorities[ca].url;
}

async function readConfigData () {
    // Read configuration file which gives
    //  1.  connection profile - that defines the blockchain network and the endpoints for its CA, Peers
    //  2.  channel name
    //  3.  chaincode name
    //  4.  wallet location - collection of certificates
    //  5.  adminUserName - identity to be used for performing transactions, initially
    //  6.  bLocahHost
    configdata = JSON.parse(fs.readFileSync('config.json', 'utf8'));

    channelName = configdata.channel_name;
    chaincodeID = configdata.chaincode_name;
    const walletLocation = configdata.wallet;
    wallet =  new FileSystemWallet(walletLocation);

    // Parse the connection profile
    const ccpPath = path.resolve(__dirname, configdata.connection_profile_filename);

    // Load connection profile; will be used to locate a gateway
    connectionProfile = JSON.parse(fs.readFileSync(ccpPath, 'utf8'));

    adminUserName = configdata.adminEnrollmentId;
    secret = configdata.enrollmentSecret;  //  same hardcoded pw for all users, for dev purposes
    bLocalHost = true;
}


/*   function setUserContext
//  Purpose:  to set the context to the user (who called this api)
//            All subsequent calls using that contract'; 'l be on this user's behalf.
//  Input:  userName - which has been registered;  If not enrolled, this function enrolls.
//  Output:  if not error returns SUCCESS else throws an exception
//          Also, (Global variable) contract will be set to this user's context
//          All further transactions using this contract will be submitted using
//          this userName
*/

async function setUserContext (userName, pwd)  {
    console.log ('>>> In function fabricClient.setUserContext: ');

    try {
        //  verify if user is registered
        await enrollUser(userName, pwd);
    } catch (error) {
        let msg = 'Verify if the user, ' + userName + ' is registered with Certificate Authority';
        console.log(msg);
        throw (prepareStatus (USER_UNAUTHORIZED, msg));
    }

    try {
        // Set connection options; identity and wallet
        let connectionOptions = {
            identity: userName,  //  * * * set context to this userName   * * *
            wallet: wallet,
            discovery: { enabled: false, asLocalhost: bLocalHost }
        };

        // Connect to gateway using application specified parameters
        console.log('Connect to Fabric gateway for userName: ' + userName);
        let userGateway = new Gateway ();
        await userGateway.connect(connectionProfile, connectionOptions);

        let network =  await userGateway.getNetwork (channelName);

        //  contract is global variable, used by invoke and query functions
        contract =  await network.getContract (chaincodeID);
        return prepareStatus(SUCCESS, 'Successfully set user context to  user:  ' + userName);
    }
    catch (error) {
        let msg = 'setUserContext failed for userName,' + userName + ' : ' + error;
        console.error(msg);
        throw (prepareStatus(FAILURE, msg));
    }
}  //  end of UserContext(userName)


function processFpcResponse (response) {
    let strResponse = response.toString();
    console.log ('Response (string): ', strResponse);
    let jsonResponse = JSON.parse(strResponse);
    // TODO: check signature/attestation ...
    let resultStr = Buffer.from(jsonResponse.ResponseData,"base64").toString();
    console.log ('Decoded ResponseData: ', resultStr);
    let result = JSON.parse(resultStr);
    if (typeof result.status === 'undefined' || typeof result.status.rc === 'undefined' || typeof result.status.message === 'undefined') {
        throw new Error("invalid response data '" + resultStr + "'");
    }
    return result;
}


//  Function to invoke transaction or do a query (evaluateTransaction)
/*
  Input:  bSubmitTransaction: true => submitTransaction will be called;
                              false => evaluateTransaction will be called;
          userName:  set caller identity
          txName, args:  parameters to submitTransaction / evaluateTransaction
*/
async function submitToFabric (bSubmitTransaction, userName, txName, ...args) {
    await setUserContext (userName,secret);

    try {
        let result;
        if (bSubmitTransaction === true) {
            result = contract.submitTransaction(txName,...args);
        }
        else {
            result = contract.evaluateTransaction(txName,...args);
        }
        return result.then((response) => {
	    try {
                return processFpcResponse(response);
            } catch (error) {
                console.log('Error on tx success return: ', error)
                throw prepareStatus(FAILURE, "Error on tx success return: "+error);
            }
        },(error) =>
        {
            try {
                // error during call to Fabric SDK, either chaincode failed or peer problem
                console.log('Error on tx return: ', error)
                // find any payload
                if (typeof error.payload === 'undefined') {
                    // if none found, just throw our backend error
                    throw new Error('No payload found');
                }
                console.log("Found payload")
                // otherwise, return (fpc) response from (fabric) payload
                return processFpcResponse(error.payload);
            } catch (error) {
                console.log('Error on tx error return: ', error)
                throw prepareStatus(FAILURE, "Error on tx error return: "+error);
            }
        });
    } catch (error) {
        // error from initial call to Fabric SDK
        console.log('Error on tx init: ', error)
        throw prepareStatus(FAILURE, "Error on tx init: "+error);
    }
}

//  Usage:  beginNetworkConnect => connectToNetwork => endNetworkConnect
//    connectToNetwork is application agnostic;
//    beginNetworkConnect and endNetworkConnect can be used for application specific functions
async function beginNetworkConnect() {
    await readConfigData();  //  from config.json

    //  enroll admin user read from config.json
    //  It is assumed that adminUserName is already registered with CA
    await enrollUser(adminUserName, secret);

}

async function endNetworkConnect () {
    //  place holder for application specific actions
    //  to be done after initial connect to network
}

//  exported functions
async function connectToNetwork() {
    console.log ('>>> Function connectToNetwork');

    await beginNetworkConnect();

    try {
        adminGateway = await new Gateway();
        await adminGateway.connect(connectionProfile,
            {
                wallet,
                identity:            adminUserName,
                discovery:           { enabled: false, localhost: bLocalHost },
                eventHandlerOptions: { strategy: null }
            });
        let network =  await adminGateway.getNetwork (channelName);
        contract =  await network.getContract (chaincodeID);
        console.log ('Connected to chaincode, ', chaincodeID);
    }  catch (error) {
        let msg = 'Error connecting to network: ' + error;
        console.log(msg);
        return (prepareStatus (FAILURE, msg));
    }

    //  Currently, no application specific code is needed in endNetworkConnect
    //  Placed here as a usage guidance
    //  await endNetworkConnect();
}

async function getRegisteredUsers() {
    let client, fabric_ca_client, idService;

    try {
        client = adminGateway.getClient();
        fabric_ca_client = client.getCertificateAuthority();
        idService = fabric_ca_client.newIdentityService();
        let adminIdentity = await adminGateway.getCurrentIdentity();

        //  adminIdentity should be a hf.Registrar
        let userList = await idService.getAll(adminIdentity);

        let  identities = userList.result.identities;

        if (identities !== undefined) {
            let users = identities.map (
                function (user) {
                    return { 'id': user.id, 'approle': getRole(user)};
                });
            return users;
        }
        else {  return [];  }
        //  TODO: the returned value is NOT in $status format.  If needed,
        //  this may be modified later to return a json as returned by prepareStatus.
    } catch (error) {
        throw (prepareStatus (FAILURE, 'Error in call to getRegisteredUsers: ' + error));
    }
}

async function enrollUser (userName, secret) {
    console.log ('>>> In function fabricClient.enrollUser: ');

    try {
        // Check to see if we have already enrolled the user.
        const userExists = await wallet.exists(userName);
        if (userExists) {
            let msg = 'An identity for the user ' + userName + ' already exists in the wallet';
            console.log(msg);
            return (prepareStatus(USER_EXISTS, msg));
        }

        // Create a new CA client for interacting with the CA.
        const caclient = new FabricCAClient(getClientCAUrl());

        const enrollment = await caclient.enroll(
            {   enrollmentID: userName,
                enrollmentSecret: secret  });

        let mspid = await getClientMspid();
        const identity = await X509WalletMixin.createIdentity(mspid,
            enrollment.certificate, enrollment.key.toBytes());
        await wallet.import(userName, identity);
        return prepareStatus(SUCCESS,
            'Successfully enrolled user ' + userName + 'and imported it into the wallet');

    } catch (error) {
        let msg = 'Failed to enroll user ' + userName + '; '  +  error;
        throw prepareStatus (USER_UNAUTHORIZED, msg);
    }
}

async function invoke(userName, txName, ...args) {
    console.log ('>>> In function, fabricClient.invoke: ');
    //  first parameter, bSubmitTransaction = true => submitTransaction will be called
    return submitToFabric (true, userName, txName, ...args);
}

async function query (userName, txName, ...args) {
    console.log ('>>> In function, fabricClient.query:  ');
    //  first parameter, bSubmitTransaction = false => evaluateTransaction will be called
    return submitToFabric (false, userName, txName, ...args);
}

module.exports = {
    connectToNetwork:connectToNetwork,
    getRegisteredUsers:getRegisteredUsers,
    invoke:invoke,
    query:query,
    enrollUser: enrollUser
};
