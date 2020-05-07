# Hello World Tutorial
This tutorial shows how to create, build, install and test  chaincode using the Fabric-Private-chaincode(FPC) framework.  This assumes familiarity with the [concepts](https://hyperledger-fabric.readthedocs.io/en/release-1.4/whatis.html#) and [the programming model](https://hyperledger-fabric.readthedocs.io/en/release-1.4/chaincode.html) in Hyperledger Fabric v1.4.


Refer to [package shim](https://godoc.org/github.com/hyperledger/fabric/core/chaincode/shim) for the GoDoc for shim interface provided by Fabric.  The FPC programming model for chaincode provides a simpler version of the _shim_ SDK provided by Hyperledger Fabric in Go and node.js.  FPC provides a C++ interface  to access its state variables and transaction context through [shim.h](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/ecc_enclave/enclave/shim.h).  The standard commands are similar to the ones in Fabric.  To ensure confidentiality of the arguments passed to the chaincode, the arguments are encrypted while using FPC SDK.

This example illustrates a simple usecase where the chaincode is used to store a single asset, `asset1` in the ledger and then retrieve the latest value of `asset1`.  Here are the steps to accomplish this:
* Develop chaincode
* Launch Fabric network
* Install and instantiate chaincode on the peer
* Init transaction
* Invoke transactions (`storeAsset` and `retrieveAsset`)
* Shut down the network

Please refer to [Architecture and
Components](https://github.com/hyperledger-labs/fabric-private-chaincode#architecture-and-components)
for more details of involved components.

## Prequisites
This tutorial presumes that the repository in https://github.com/hyperledger-labs/fabric-private-chaincode has been installed as per [README.md](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/README.md#requirements) in `FPC-INSTALL-DIR` which is in the `$GOPATH` folder.

## Develop chaincode
* Create a folder named `helloworld`  in `FPC-INSTALL-DIR/examples`.
```
cd $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode/examples
mkdir helloworld
cd helloworld
touch helloworld_cc.cpp
```

* Add the necessary includes and a preliminary version of `init` function.  A chaincode has to implement the `init` method which the peer calls when the chaincode is instantiated. The result of the transaction is returned in `response`.  `ctx` represents the transaction context.  The function body also illustrates the use of context, e.g., it allows us to retrieve the invocation arguments as defined, e.g., with `--ctor` parameter of `chaincode instantiate` peer cli command.

```
#include "shim.h"
#include "logging.h"
#include <string>

// implements chaincode logic for init
int init(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    std::vector<std::string> argss;
    get_string_args(argss, ctx);

    return 0;
}
```

* Add preliminary version of `invoke` function.  The arguments have the same semantics as for `init`. The function body illustrates another way to get invocation parameters, similar to functions provided in go shim.


```

// implements chaincode logic for invoke
int invoke(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);

    return 0;
}
```

Let us add the first transaction, `storeAsset` which simply saves the value of an asset by calling `put_state` method defined in shim.h.  LOG_DEBUG sends log messages to the file `/tmp/hyperledger/test/peer.err` (if the environment variable `SGX_BUILD` is set to `DEBUG`).
```
#define OK "OK"

//  Add asset_name, value to ledger
std::string storeAsset(std::string asset_name, int value, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("HelloworldCC: +++ storeAsset +++");

    put_state(asset_name.c_str(), (uint8_t*)&value, sizeof(int), ctx);

    return OK;
}
```


Similarly, let us add the next transaction, `retrieveAsset` which reads the value of an asset by calling `get_state` method defined in shim.h.


```
#define NOT_FOUND "Asset not found"
#define MAX_VALUE_SIZE 1024

//  Get value set for asset_name in ledger
std::string retrieveAsset(std::string asset_name, shim_ctx_ptr_t ctx)
{
    std::string result;
    LOG_DEBUG("HelloworldCC: +++ retrieveAsset +++");

    uint32_t asset_bytes_len = 0;
    uint8_t asset_bytes[MAX_VALUE_SIZE];
    get_state(asset_name.c_str(), asset_bytes, sizeof(asset_bytes), &asset_bytes_len, ctx);

    //  check if asset_name exists
    if (asset_bytes_len > 0)
    {
        //  asset exists;  return value
        result = asset_name + ":" + std::to_string((int)(*asset_bytes));
    }
    else
    {
        //  asset does not exist
        result = NOT_FOUND;
    }
    return result;
}
```

Modify the `invoke` method to invoke the appropriate function depending on the function name passed in `args` and return `response`.
```
#define OK "OK"
#define NOT_FOUND "Asset not found"

#define MAX_VALUE_SIZE 1024

// implements chaincode logic for invoke
int invoke(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("HelloworldCC: +++ Executing helloworld chaincode invocation +++");
    LOG_DEBUG("HelloworldCC: \tArgs: %s", args);

    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);
    std::string asset_name = params[0];
    std::string result;

    if (function_name == "storeAsset")
    {
        int value = std::stoi (params[1]);
        result = storeAsset(asset_name, value, ctx);
    }
    else if (function_name == "retrieveAsset")
    {
        result = retrieveAsset(asset_name, ctx);
    }
    else
    {
        // unknown function
        LOG_DEBUG("HelloworldCC: RECEIVED UNKNOWN transaction '%s'", function_name);
        return -1;
    }
     // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // error:  buffer too small for the response to be sent
        LOG_DEBUG("HelloworldCC: Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }

    // copy result to response
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("HelloworldCC: Response: %s", result.c_str());
    LOG_DEBUG("HelloworldCC: +++ Executing done +++");
    return 0;
}
```

Let us also modify `init` method to retrieve and log who triggered the transaction.
```
// implements chaincode logic for init
int init(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    char real_bidder_name_msp_id[1024];
    char real_bidder_name_dn[1024];
    get_creator_name(real_bidder_name_msp_id, sizeof(real_bidder_name_msp_id),
        real_bidder_name_dn, sizeof(real_bidder_name_dn), ctx);

    LOG_DEBUG("HelloworldCC: +++ Chaincode is initialized by (msp_id: %s, dn: %s) +++",
        real_bidder_name_msp_id, real_bidder_name_dn);

    return 0;
}
```

Here is the complete file, `helloworld_cc.cpp`:
```
#include "shim.h"
#include "logging.h"
#include <string>

#define OK "OK"
#define NOT_FOUND "Asset not found"

#define MAX_VALUE_SIZE 1024


// implements chaincode logic for init
int init(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    char real_bidder_name_msp_id[1024];
    char real_bidder_name_dn[1024];
    get_creator_name(real_bidder_name_msp_id, sizeof(real_bidder_name_msp_id),
        real_bidder_name_dn, sizeof(real_bidder_name_dn), ctx);

    LOG_DEBUG("HelloworldCC: +++ Chaincode is initialized by (msp_id: %s, dn: %s) +++",
        real_bidder_name_msp_id, real_bidder_name_dn);

    return 0;
}

//  Add asset_name, value to ledger
std::string storeAsset(std::string asset_name, int value, shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("HelloworldCC: +++ storeAsset +++");

    put_state(asset_name.c_str(), (uint8_t*)&value, sizeof(int), ctx);

    return OK;
}

std::string retrieveAsset(std::string asset_name, shim_ctx_ptr_t ctx)
{
    std::string result;
    LOG_DEBUG("HelloworldCC: +++ retrieveAsset +++");

    uint32_t asset_bytes_len = 0;
    uint8_t asset_bytes[MAX_VALUE_SIZE];
    get_state(asset_name.c_str(), asset_bytes, sizeof(asset_bytes), &asset_bytes_len, ctx);

    //  check if asset_name exists
    if (asset_bytes_len > 0)
    {
        result = asset_name +   ":" +  std::to_string((int)(*asset_bytes));
     } else {
        //  asset does not exist
        result = NOT_FOUND;
    }
    return result;
}

// implements chaincode logic for invoke
int invoke(
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    shim_ctx_ptr_t ctx)
{
    LOG_DEBUG("HelloworldCC: +++ Executing helloworld chaincode invocation +++");

    std::string function_name;
    std::vector<std::string> params;
    get_func_and_params(function_name, params, ctx);
    std::string asset_name = params[0];
    std::string result;

    if (function_name == "storeAsset")
    {
        int value = std::stoi (params[1]);
        result = storeAsset(asset_name, value, ctx);
    }
    else if (function_name == "retrieveAsset")
    {
        result = retrieveAsset(asset_name, ctx);
    }
    else
    {
        // unknown function
        LOG_DEBUG("HelloworldCC: RECEIVED UNKNOWN transaction '%s'", function_name);
        return -1;
    }

    // check that result fits into response
    int neededSize = result.size();
    if (max_response_len < neededSize)
    {
        // error:  buffer too small for the response to be sent
        LOG_DEBUG("HelloworldCC: Response buffer too small");
        *actual_response_len = 0;
        return -1;
    }

    // copy result to response
    memcpy(response, result.c_str(), neededSize);
    *actual_response_len = neededSize;
    LOG_DEBUG("HelloworldCC: Response: %s", result.c_str());
    LOG_DEBUG("HelloworldCC: +++ Executing done +++");
    return 0;
}
```

## Build

Make sure you have the [environment variables](https://github.com/hyperledger-labs/fabric-private-chaincode#environment-settings) set.  In addition, set `SGX_BUILD=DEBUG` to enable log messages.

To build the helloworld chaincode, we are using CMake. This simplifies the build process and compiles our chaincode using the SGX SDK. Create `CMakeLists.txt` with the following content.

File: `CMakeLists.txt`
```
cmake_minimum_required(VERSION 3.5.1)

set(SOURCE_FILES
    helloworld_cc.cpp
    )

include(../../ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)
```

Create `Makefile` with the following content.  For your convenience, you can copy the `Makefile` from `FPC-INSTALL-DIR/examples/auction` folder.
File: `examples/helloworld/Makefile`
```
TOP = ../..
include $(TOP)/build.mk

BUILD_DIR := _build

$(BUILD_DIR):
	@if [ ! -d $(BUILD_DIR) ]; then \
		mkdir -p $(BUILD_DIR) && \
		cd $(BUILD_DIR) && \
		cmake ./..; \
	fi

build: $(BUILD_DIR)
	$(MAKE) --directory=$<

clean:
	rm -rf $(BUILD_DIR)
```


In `FPC-INSTALL-DIR/examples/helloworld` folder, to build the chaincode, execute:
```
make
```

Following is a part of expected output.  Please note `[100%] Built target enclave` message in the output.  This suggests that build was successful.
Output:
```
make[3]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
[100%] Built target enclave
make[2]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
/usr/bin/cmake -E cmake_progress_start /home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build/CMakeFiles 0
make[1]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
```


## Time to test!
Next step is to test the chaincode by invoking the transactions, for which you need a basic Fabric network with a channel. You will use the FPC test framework to bring up a Fabric network in which the helloworld code can be executed as a chaincode _in an SGX enclave_.  The Fabric network used in this tutorial is defined and configured using `integration/config/core.yaml`.  Specifically, please note the additions to the standard Fabric configurations.  These are marked as `FPC Addition`;  these enable the integration points with Fabric and have to be replicated if you want to use your own Fabric configuration.

Create a file `test.sh` in `examples/helloworld` folder as follows.  Note that the initial lines in the script points to files and folders in FPC framework.
-`FPC_TOP_DIR` points to FPC-INSTALL-DIR
-`FABRIC_CFG_PATH` points to the FPC-INSTALL-DIR/integration/config, which contains yaml files that define the Fabric network
-`FABRIC_SCRIPTDIR` points to scripts with custom FPC wrappers and utility scripts.

File:  test.sh
```
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../.."
FABRIC_CFG_PATH="${SCRIPTDIR}/../../integration/config"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=helloworld_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/helloworld/_build/lib/

CC_VERS=0
```

### Launch Fabric network

Now that the environment is set, add commands to clear docker containers previously created, if any.   Add calls to `ledger_init` (which sets up a test network), run helloworld test and `ledger_shutdown` (which cleans up the test network).
```
# 1. prepare
para
say "Preparing Helloworld Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

# 2. run
say "- setup ledger"
ledger_init

say "- helloworld test"
helloworld_test    #  yet to be created

say "- shutdown ledger"
ledger_shutdown
```

### Install and instantiate chaincode on the peer
Like in the case of Fabric, you install, instantiate the chaincode and then invoke transactions.  To install a chaincode, you need to use `FPC-INSTALL-DIR/fabric/bin/peer.sh`.  This is a custom FPC wrapper to be used _instead_ of the `peer` cli command from Fabric.  `${PEER_CMD}` is set in `FPC-INSTALL-DIR/fabric/bin/lib/common_ledger.sh` and conveniently points to the required script file.
With the variables set and `common_ledger.sh` executed, usage of `peer.sh` is as follows:
```
try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
```

Add the following content to the function, `helloworld_test()` in test.sh.  Please note the inline comments for each of the commands.

```
helloworld_test() {
    say "- do hello world"
    # install and instantiate helloworld  chaincode

    # builds the docker image; creates the docker container and enclave;
    # input:  CC_ID:chaincode name; CC_VERS:chaincode version;
    #         ENCLAVE_SO_PATH:path to build artifacts
    say "- install helloworld chaincode"
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    # instantiate helloworld chaincode
    say "- instantiate helloworld chaincode"
    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    # store the value of 100 in asset1
    say "- invoke storeAsset transaction to store value 100 in asset1"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["storeAsset","asset1","100"]}' --waitForEvent

    # retrieve current value for "asset1";  should be 100;
    say "- invoke retrieveAsset transaction to retrieve current value of asset1"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["retrieveAsset","asset1"]}' --waitForEvent

    say "- invoke query with retrieveAsset transaction to retrieve current value of asset1"
    try ${PEER_CMD} chaincode query  -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["retrieveAsset","asset1"]}'
}
```

Putting all these code snippets together, here is the complete `test.sh` file.
```
#!/bin/bash

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/../.."
FABRIC_CFG_PATH="${SCRIPTDIR}/../../integration/config"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=helloworld_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/helloworld/_build/lib/

CC_VERS=0

helloworld_test() {
    say "- do hello world"
    # install and instantiate helloworld  chaincode

    # builds the docker image; creates the docker container and enclave;
    # input:  CC_ID:chaincode name; CC_VERS:chaincode version;
    #         ENCLAVE_SO_PATH:path to build artifacts

    say "- install helloworld chaincode"
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    # instantiate helloworld chaincode
    say "- instantiate helloworld chaincode"
    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    # store the value of 100 in asset1
    say "- invoke storeAsset transaction to store value 100 in asset1"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["storeAsset","asset1","100"]}' --waitForEvent

    # retrieve current value for "asset1";  should be 100;
    say "- invoke retrieveAsset transaction to retrieve current value of asset1"
    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["retrieveAsset","asset1"]}' --waitForEvent

    say "- invoke query with retrieveAsset transaction to retrieve current value of asset1"
    try ${PEER_CMD} chaincode query  -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["retrieveAsset","asset1"]}'
}

# 1. prepare
para
say "Preparing Helloworld Test ..."
# - clean up relevant docker images
docker_clean ${ERCC_ID}
docker_clean ${CC_ID}

trap ledger_shutdown EXIT

para
say "Run helloworld  test"

say "- setup ledger"
ledger_init

say "- helloworld test"
helloworld_test

say "- shutdown ledger"
ledger_shutdown

para
yell "Helloworld test PASSED"

exit 0
```


Assuming we are still in `$GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode/examples`, execute the test script:
```
cd $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld
bash ./test.sh
```

If you see the message, `test.sh: Helloworld test PASSED`, then the transactions have been successfully executed.  Now, let us look the responses from the individual trasactions.


Output from test.sh for `storeAsset` transaction invocation will look like:
```
test.sh: - invoke storeAsset transaction to store value 100 in asset1
2019-08-25 22:47:07.022 UTC [chaincodeCmd] ClientWait -> INFO 001 txid [02b87c6451eab6ba2f5b9a80c6a80898a50a1adc20e94d79e91d7f8ae7a906c0] committed with status (VALID) at
2019-08-25 22:47:07.022 UTC [chaincodeCmd] chaincodeInvokeOrQuery -> INFO 002 Chaincode invoke successful. result: status:200 payload:"{\"ResponseData\":\"T0s=\",\"Signature\":\"MEUCIQDF1uNxDKEx2KftztrgB4GfN8xJ9jq47NU/NgRRM0EqswIgWk63ZwiamGmFfOHy+C11I3a2z4831t0Y2qrzhVLJv9g=\",\"PublicKey\":\"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExQZkgJymGfDZC7JcgjbacJdpX6Bbb8YvBUjX7Su3T1SJRALbJ2fSppIbHrXjmG1y6MQN41OYGKV/FhNA8FPWdQ==\"}"
```

Response from the transaction is:
```
{
	"ResponseData": "T0s=",
	"Signature": "MEUCIQDF1uNxDKEx2KftztrgB4GfN8xJ9jq47NU/NgRRM0EqswIgWk63ZwiamGmFfOHy+C11I3a2z4831t0Y2qrzhVLJv9g=",
	"PublicKey": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExQZkgJymGfDZC7JcgjbacJdpX6Bbb8YvBUjX7Su3T1SJRALbJ2fSppIbHrXjmG1y6MQN41OYGKV/FhNA8FPWdQ=="
}
```

Response from the transactions is marked by `ResponseData`.  In this case, "ResponseData" is _Base64 encoded string_ of "OK".  In addition, the response also contains the signature of the peer and the public key of the enclave in which the chaincode was executed.

Use `base64` to decode "ResponseData" i.e. `T0s=`:
```
$ echo "T0s=" | base64 -D
OK
```

Let us look at the output of `retrieveAsset` transaction.
```
test.sh: - invoke retrieveAsset transaction to retrieve current value of asset1
2019-08-25 22:47:09.292 UTC [chaincodeCmd] ClientWait -> INFO 001 txid [530658f431d04511e2eecee7b7582bb3c6b449f99e9845c6d1327e773f756f13] committed with status (VALID) at
2019-08-25 22:47:09.293 UTC [chaincodeCmd] chaincodeInvokeOrQuery -> INFO 002 Chaincode invoke successful. result: status:200 payload:"{\"ResponseData\":\"YXNzZXQxOjEwMA==\",\"Signature\":\"MEQCIEiEdmV5jGW9kd2+a4QEeoEIlNFYB3NCvA+xtXKEkzK9AiAxIVPo0ca0s39KyY9N7tdd5ZfWXN5eSf+KeBSQPMaHKA==\",\"PublicKey\":\"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExQZkgJymGfDZC7JcgjbacJdpX6Bbb8YvBUjX7Su3T1SJRALbJ2fSppIbHrXjmG1y6MQN41OYGKV/FhNA8FPWdQ==\"}"
```

Response from the transaction is:
```
{
	"ResponseData": "YXNzZXQxOjEwMA==",
	"Signature": "MEQCIEiEdmV5jGW9kd2+a4QEeoEIlNFYB3NCvA+xtXKEkzK9AiAxIVPo0ca0s39KyY9N7tdd5ZfWXN5eSf+KeBSQPMaHKA==",
	"PublicKey": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAExQZkgJymGfDZC7JcgjbacJdpX6Bbb8YvBUjX7Su3T1SJRALbJ2fSppIbHrXjmG1y6MQN41OYGKV/FhNA8FPWdQ=="
}
```

Use `base64` to decode "ResponseData" i.e. `YXNzZXQxOjEwMA==` which is _Base64 encoded string_ of "asset1:100".
```
$ echo "YXNzZXQxOjEwMA==" | base64 -D
asset1:100
```


Output of `retrieveAsset` query will look like:
```
test.sh: - invoke query with retrieveAsset transaction to retrieve current value of asset1
{"ResponseData":"YXNzZXQxOjEwMA==","Signature":"MEQCIHViPWLKQvd2RezK8oKk4E9nY1WrAR1F5J0wFptX4erVAiAZgTmug8QBqZaFWLPBCfRFWste66H7QN5a3BOA+G5aCg==","PublicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkx5nDA1TjDA5j4b8cmQcU0JpjALu6UFZKNOpttRJCNlaAighZi8ftnjJTeMuyHgEQHeHz/m/p/noJcxilrtMxQ=="}
```


Response from the query is:
```
{
    "ResponseData":"YXNzZXQxOjEwMA==",
    "Signature":"MEQCIHViPWLKQvd2RezK8oKk4E9nY1WrAR1F5J0wFptX4erVAiAZgTmug8QBqZaFWLPBCfRFWste66H7QN5a3BOA+G5aCg==",
    "PublicKey": "MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEkx5nDA1TjDA5j4b8cmQcU0JpjALu6UFZKNOpttRJCNlaAighZi8ftnjJTeMuyHgEQHeHz/m/p/noJcxilrtMxQ=="
}
```

The response from the query is the same as the corresponding transaction.  "ResponseData" is _Base64 encoded string_ of "asset1:100".

Yay !  You did it !

### Interactive testing

If you want to interactively test FPC, you can use the commands in `../fabric/bin/`. It contains the standard cli commands you expect from fabric -- note: it is though important that you use these scripts rather than use the fabric commands directly. The provide the same interface as the fabric commands but do some additional magic under the cover.   Additionally, there are also two convenience functions for quickly setting up and shutting down a ledger: `ledger_init.sh` and `ledger_shutdown.sh`.
