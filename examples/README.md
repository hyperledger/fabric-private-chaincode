# Hello World Tutorial
This tutorial shows how to create, build, install and test  chaincode using the Fabric-Private-chaincode(FPC) framework.  This assumes familiarity with the [concepts](https://hyperledger-fabric.readthedocs.io/en/release-2.0/whatis.html#) and [the programming model](https://hyperledger-fabric.readthedocs.io/en/release-2.0/chaincode.html) in Hyperledger Fabric v2.0.


Refer to [package shim](https://pkg.go.dev/github.com/hyperledger/fabric-chaincode-go/shim?tab=doc) for the GoDoc for shim interface provided by Fabric.  The FPC programming model for chaincode provides a simpler version of the _shim_ SDK provided by Hyperledger Fabric in Go and node.js.  FPC provides a C++ interface  to access its state variables and transaction context through [shim.h](../ecc_enclave/enclave/shim.h).  The standard commands are similar to the ones in Fabric.  To ensure confidentiality of the arguments passed to the chaincode, the arguments are transparently encrypted while using FPC SDK.

Regarding management functionally such as chaincode installation and alike, plus refer to  [FPC Management API document](../docs/design/fabric-v2%2B/fpc-management.md) for details.

This example illustrates a simple usecase where the chaincode is used to store a single asset, `asset1` in the ledger and then retrieve the latest value of `asset1`.  Here are the steps to accomplish this:
* Develop chaincode
* Launch Fabric network
* Install and instantiate chaincode on the peer
* Invoke transactions (`storeAsset` and `retrieveAsset`)
* Shut down the network

Please refer to [Architecture and
Components](https://github.com/hyperledger-labs/fabric-private-chaincode#architecture-and-components)
for more details of involved components.

## Prequisites
This tutorial presumes that the repository in https://github.com/hyperledger-labs/fabric-private-chaincode has been installed as per [README.md](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/README.md#requirements) in `FPC-INSTALL-DIR` which is in the `$GOPATH` folder.

## Develop chaincode
* Create a folder named `helloworld`  in `FPC-INSTALL-DIR/examples`.
```bash
cd $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode/examples
mkdir helloworld
cd helloworld
touch helloworld_cc.cpp
```

* Add the necessary includes and a preliminary version of `invoke` function. The result of the transaction is returned in `response`. `ctx` represents the transaction context. The function body illustrates another way to get invocation parameters, similar to functions provided in go shim.
```c++
#include "shim.h"
#include "logging.h"
#include <string>

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

Let us add the first transaction, `storeAsset` which simply saves the value of an asset by calling `put_state` method defined in `shim.h`.
`LOG_DEBUG` sends log messages to the file `/tmp/fpc-extbuilder.${date_time}.{chaincode_name}/chaincode.log` (if the environment variable `SGX_BUILD` is set to `DEBUG`).
```c++
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


```c++
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
```c++
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

Here is the complete file, `helloworld_cc.cpp`:
```c++
#include "shim.h"
#include "logging.h"
#include <string>

#define OK "OK"
#define NOT_FOUND "Asset not found"

#define MAX_VALUE_SIZE 1024

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
```CMake
cmake_minimum_required(VERSION 3.5.1)

set(SOURCE_FILES
    helloworld_cc.cpp
    )

include(../../ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)
```

Create `Makefile` with the following content.  For your convenience, you can copy the `Makefile` from `FPC-INSTALL-DIR/examples/auction` folder.
File: `examples/helloworld/Makefile`
```Makefile
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
```bash
make
```

Following is a part of expected output.  Please note `[100%] Built target enclave` message in the output.  This suggests that build was successful.
Output:
```bash
make[3]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
[100%] Built target enclave
make[2]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
/usr/bin/cmake -E cmake_progress_start /home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build/CMakeFiles 0
make[1]: Leaving directory '/home/bcuser/work/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld/_build'
```


## Time to test!
Next step is to test the chaincode by invoking the transactions, for which you need a basic Fabric network with a channel. You will use the FPC test framework to bring up a Fabric network in which the helloworld code can be executed as a chaincode _in an SGX enclave_.  The Fabric network used in this tutorial is defined and configured using `integration/config/core.yaml`.  Specifically, please note the additions to the standard Fabric configurations.  These are marked as `FPC Addition`;  these enable the integration points with Fabric and have to be replicated if you want to use your own Fabric configuration.

Create a file `test.sh` in `examples/helloworld` folder as follows.  Note that the initial lines in the script points to files and folders in FPC framework.
-`FPC_PATH` points to FPC-INSTALL-DIR
-`FABRIC_CFG_PATH` points to the FPC-INSTALL-DIR/integration/config, which contains yaml files that define the Fabric network
-`FABRIC_SCRIPTDIR` points to scripts with custom FPC wrappers and utility scripts.

File:  test.sh
```Makefile
SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/../.."
FABRIC_CFG_PATH="${SCRIPTDIR}/../../integration/config"
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

#this is the path that will be used for the docker build of the chaincode enclave
CC_PATH=${FPC_PATH}/examples/helloworld/_build/lib/

CC_ID=helloworld_test
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_EP="OR('SampleOrg.member')"
CC_SEQ="1"
```

### Launch Fabric network

Now that the environment is set, add commands to clear docker containers previously created, if any.   Add calls to `ledger_init` (which sets up a test network), run helloworld test and `ledger_shutdown` (which cleans up the test network).
```bash
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

### Install our FPC Chaincode
Like in the case of Fabric, you install the chaincode using the `peer lifecycle` commands and then invoke transactions. 
To install the FPC chaincode, you need to use `FPC-INSTALL-DIR/fabric/bin/peer.sh`.  This is a custom FPC wrapper to be used _instead_ of the `peer` cli command from Fabric.  `${PEER_CMD}` is set in `FPC-INSTALL-DIR/fabric/bin/lib/common_ledger.sh` and conveniently points to the required script file.
With the variables set and `common_ledger.sh` executed, usage of `peer.sh` is as follows:
```bash
    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label ${CC_ID} --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}
```

In the next step, the FPC chaincode must be approved by the organizations on the channel by agreeing on the chaincode definition.
```bash
    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
```

To complete the installation, we need to create an enclave that runs the FPC Chaincode.
```bash
    # create an FPC Chaincode enclave
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}
```

Add the following content to the function, `helloworld_test()` in test.sh.  Please note the inline comments for each of the commands.

```bash
helloworld_test() {
    say "- do hello world"

    # install helloworld  chaincode
    # input:  CC_ID:chaincode name; CC_VER:chaincode version;
    #         CC_PATH:path to build artifacts
    say "- install helloworld chaincode"
    PKG=/tmp/${CC_ID}.tar.gz
    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label ${CC_ID} --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}

    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    # create an FPC Chaincode enclave
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}

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
```bash
#!/bin/bash

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_PATH="${SCRIPTDIR}/../.."
FABRIC_CFG_PATH="${SCRIPTDIR}/../../integration/config"
FABRIC_SCRIPTDIR="${FPC_PATH}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

#this is the path that will be used for the docker build of the chaincode enclave
CC_PATH=${FPC_PATH}/examples/helloworld/_build/lib/

CC_ID=helloworld_test
CC_VER="$(cat ${CC_PATH}/mrenclave)"
CC_EP="OR('SampleOrg.member')"
CC_SEQ="1"

helloworld_test() {
    say "- do hello world"

    # install helloworld  chaincode
    # input:  CC_ID:chaincode name; CC_VER:chaincode version;
    #         CC_PATH:path to build artifacts
    say "- install helloworld chaincode"
    PKG=/tmp/${CC_ID}.tar.gz
    try ${PEER_CMD} lifecycle chaincode package --lang fpc-c --label ${CC_ID} --path ${CC_PATH} ${PKG}
    try ${PEER_CMD} lifecycle chaincode install ${PKG}

    PKG_ID=$(${PEER_CMD} lifecycle chaincode queryinstalled | awk "/Package ID: ${CC_ID}/{print}" | sed -n 's/^Package ID: //; s/, Label:.*$//;p')

    try ${PEER_CMD} lifecycle chaincode approveformyorg -o ${ORDERER_ADDR} -C ${CHAN_ID} --package-id ${PKG_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode checkcommitreadiness -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}
    try ${PEER_CMD} lifecycle chaincode commit -o ${ORDERER_ADDR} -C ${CHAN_ID} --name ${CC_ID} --version ${CC_VER} --sequence ${CC_SEQ} --signature-policy ${CC_EP}

    # create an FPC Chaincode enclave
    try ${PEER_CMD} lifecycle chaincode initEnclave -o ${ORDERER_ADDR} --peerAddresses "localhost:7051" --name ${CC_ID}

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
```hash
cd $GOPATH/src/github.com/hyperledger-labs/fabric-private-chaincode/examples/helloworld
bash ./test.sh
```

If you see the message, `test.sh: Helloworld test PASSED`, then the transactions have been successfully executed.  Now, let us look the responses from the individual trasactions.


Output from test.sh for `storeAsset` transaction invocation will look like:
```bash
test.sh: - invoke storeAsset transaction to store value 100 in asset1
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query -o localhost:7050 -C mychannel -n ercc -c {"Args":["QueryChaincodeEncryptionKey", "helloworld_test"]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query -o localhost:7050 -C mychannel -n ercc -c {"Args":["QueryChaincodeEndPoints", "helloworld_test"]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query --peerAddresses localhost:7051 -o localhost:7050 -C mychannel -n helloworld_test -c {"Args":["__invoke", "CoACqxRdrnlNKNVavdNKiMD5jJn+p4goTugbfVyQMPZ6qQxaBFoqiXpSW6ubzj1d5yutxdyP9oDdbrRdEcMyuOcTVzx94mPpSLHEF+V6Hm+6VMbXE2M3JBgzP9U0kIrrgrekwOYVfXDlnWex9oSWSdLL2rGZJhBsODvZ/1K99Ey7X+8cHe9wTEnLT/RpcVjWrnxpXiJ/ZmfascANV5MRkRsyT1LAEZUDK8u6vhwdLLGbbXy5UmnLCqEMdrPEWHJNrULiKfRTUdUx9NrqwRyZgS0ZDL95MrRhEtgtE3gkP5olslFn5qm2TRSXh02CdTh5w94vOC55dK+BtZW2l3ckPS1UMQ=="]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode invoke -o localhost:7050 --waitForEvent -C mychannel -n helloworld_test -c {"Args":["__endorse", "CvoMCiAAGpKnvy4pnwKl/9aWElKT10xcM+5+O02AANy6OItihRI6CjgaNgoGYXNzZXQxGix3UVowT3prS0JmeHBDWWw2aXdGNlNEV2FKMWd2Z0drVVczLzA5a0duYm13PRq1CwrpCgrcBwpyCAMaDAjf+PaABhDI1OKsASIJbXljaGFubmVsKkBlOWY5Yjc0MWQzYjgxYTNhNzA2ZTVjNDQ1MDcyNzdiZjU1MzRlNDYxNWRjMmIwMjMzMmU4MWZlMTZjMTYwY2EwOhMSERIPaGVsbG93b3JsZF90ZXN0EuUGCsgGCglTYW1wbGVPcmcSugYtLS0tLUJFR0lOIENFUlRJRklDQVRFLS0tLS0KTUlJQ05qQ0NBZDJnQXdJQkFnSVJBTW5mOS9kbVY5UnZDQ1Z3OXBaUVVmVXdDZ1lJS29aSXpqMEVBd0l3Z1lFeApDekFKQmdOVkJBWVRBbFZUTVJNd0VRWURWUVFJRXdwRFlXeHBabTl5Ym1saE1SWXdGQVlEVlFRSEV3MVRZVzRnClJuSmhibU5wYzJOdk1Sa3dGd1lEVlFRS0V4QnZjbWN4TG1WNFlXMXdiR1V1WTI5dE1Rd3dDZ1lEVlFRTEV3TkQKVDFBeEhEQWFCZ05WQkFNVEUyTmhMbTl5WnpFdVpYaGhiWEJzWlM1amIyMHdIaGNOTVRjeE1URXlNVE0wTVRFeApXaGNOTWpjeE1URXdNVE0wTVRFeFdqQnBNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0JNS1EyRnNhV1p2CmNtNXBZVEVXTUJRR0ExVUVCeE1OVTJGdUlFWnlZVzVqYVhOamJ6RU1NQW9HQTFVRUN4TURRMDlRTVI4d0hRWUQKVlFRREV4WndaV1Z5TUM1dmNtY3hMbVY0WVcxd2JHVXVZMjl0TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRVo4UzRWNzFPQkpweU1JVlpkd1lkRlhBY2tJdHJwdlNyQ2YwSFFnNDBXVzlYU29PT083NkkrVW1mCkVrbVRsSUpYUDcvQXlSUlNSVTM4b0k4SXZ0dTRNNk5OTUVzd0RnWURWUjBQQVFIL0JBUURBZ2VBTUF3R0ExVWQKRXdFQi93UUNNQUF3S3dZRFZSMGpCQ1F3SW9BZ2luT1JJaG5QRUZaVWhYbTZlV0JrbTdLN1pjOFI0L3o3TFc0SApvc3NEbENzd0NnWUlLb1pJemowRUF3SURSd0F3UkFJZ1Zpa0lVWnpnZnVGc0dMUUhXSlVWSkNVN3BEYUVUa2F6ClB6RmdzQ2lMeFVBQ0lDZ3pKWWxXN252WnhQN2I2dGJldTN0OG1yaE1YUXM5NTZtRDQrQm9LdU5JCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KEhgBT3hTOjs8cnlaAk1LwePi40qsObkDW04ShwMKhAMKgQMIARIREg9oZWxsb3dvcmxkX3Rlc3Qa6QIKCF9faW52b2tlCtwCQ29BQ3F4UmRybmxOS05WYXZkTktpTUQ1akpuK3A0Z29UdWdiZlZ5UU1QWjZxUXhhQkZvcWlYcFNXNnViemoxZDV5dXR4ZHlQOW9EZGJyUmRFY015dU9jVFZ6eDk0bVBwU0xIRUYrVjZIbSs2Vk1iWEUyTTNKQmd6UDlVMGtJcnJncmVrd09ZVmZYRGxuV2V4OW9TV1NkTEwyckdaSmhCc09EdlovMUs5OUV5N1grOGNIZTl3VEVuTFQvUnBjVmpXcm54cFhpSi9abWZhc2NBTlY1TVJrUnN5VDFMQUVaVURLOHU2dmh3ZExMR2JiWHk1VW1uTENxRU1kclBFV0hKTnJVTGlLZlJUVWRVeDlOcnF3UnlaZ1MwWkRMOTVNclJoRXRndEUzZ2tQNW9sc2xGbjVxbTJUUlNYaDAyQ2RUaDV3OTR2T0M1NWRLK0J0WlcybDNja1BTMVVNUT09EkcwRQIhANz3BNlcghQhaGzN+Z7rem/JhcDXL8h2y3i4F7vFF7MBAiATGtz3GxXtCQZqJ2cjtOjhLChhkcqyIGclIQgC4W7JpSIgUCM/JIUzwZ7qRQXb1i/NmHUhccYDMQEqYBwa7KTfttcqQEVCNjA5M0ZCRTE4MTk0MTRGRUU3MzUxODAyRjYyNkE5NjY3OTFBNkQ4OTQzMkIwRTJCNURGNDBFOTNEMUI0Q0YSRjBEAiAtX4zseT2meebl5vLdCnlt5hLnH+QA3U4yAa9x7YuhEgIgY0LkRSBJkA3SW7Rpj27dXLvSmVd8kyK8c2Srqz/ZhIQ="]}

```

Response from the transaction is:
```
OK
```

Let us look at the output of `retrieveAsset` transaction.
```bash
test.sh: - invoke retrieveAsset transaction to retrieve current value of asset1
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query -o localhost:7050 -C mychannel -n ercc -c {"Args":["QueryChaincodeEncryptionKey", "helloworld_test"]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query -o localhost:7050 -C mychannel -n ercc -c {"Args":["QueryChaincodeEndPoints", "helloworld_test"]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode query --peerAddresses localhost:7051 -o localhost:7050 -C mychannel -n helloworld_test -c {"Args":["__invoke", "CoACGbE4bFpQaubzAw/XQPRJnuD2bC3ZcCV4b9TC0gmrPIe+YBYMHYb5WDhNAPrgJvSCLiJbXb15qzp1C0I0lhSVadVLH465CL5huLbDwtmLswpGDVlswHPHmik+ce7Xx04hEqNupu3VUxar6dUdcUh8wyyW2fgj/q1f96ZAm4SH/DAjdu7y3qSqIlPp1LMpz+7SpM6/AVAP3aSuDetooyTitDjYMHU7OUxkj2pH41MSqA0g65/bszPQXKbcc49FivwnKEKyIV8KyXVrPB1s8JkpP89HfCQxmkRU3pmukMi4jkLxYT4tXz4MCwbJJp4K6aleCaz1sjdHONcIBveFb/nekQ=="]}
peer.sh: /project/src/github.com/hyperledger/fabric/build/bin/peer chaincode invoke -o localhost:7050 --waitForEvent -C mychannel -n helloworld_test -c {"Args":["__endorse", "CvkMCiyEzRpNIXSgFVu14waJqpDwF6JIKQy0sLD2hoPjt3/Qhayc6Im3VCF1FCboqhIuCgoKCAoGYXNzZXQxEiBu8sguYb+l8QfD8ToPBdPqoJbiZhYt/6AaZvCDqo47ARq0CwrpCgrcBwpyCAMaDAji+PaABhC4hKWDAyIJbXljaGFubmVsKkBkNTVkYmQwNmNmZWUxMmJhYzczNjc1MzlkMTFmNGUxY2QyOGZjNzNlNTJmMjczYjMwMDkzYWMxNTBhOWFhNDVlOhMSERIPaGVsbG93b3JsZF90ZXN0EuUGCsgGCglTYW1wbGVPcmcSugYtLS0tLUJFR0lOIENFUlRJRklDQVRFLS0tLS0KTUlJQ05qQ0NBZDJnQXdJQkFnSVJBTW5mOS9kbVY5UnZDQ1Z3OXBaUVVmVXdDZ1lJS29aSXpqMEVBd0l3Z1lFeApDekFKQmdOVkJBWVRBbFZUTVJNd0VRWURWUVFJRXdwRFlXeHBabTl5Ym1saE1SWXdGQVlEVlFRSEV3MVRZVzRnClJuSmhibU5wYzJOdk1Sa3dGd1lEVlFRS0V4QnZjbWN4TG1WNFlXMXdiR1V1WTI5dE1Rd3dDZ1lEVlFRTEV3TkQKVDFBeEhEQWFCZ05WQkFNVEUyTmhMbTl5WnpFdVpYaGhiWEJzWlM1amIyMHdIaGNOTVRjeE1URXlNVE0wTVRFeApXaGNOTWpjeE1URXdNVE0wTVRFeFdqQnBNUXN3Q1FZRFZRUUdFd0pWVXpFVE1CRUdBMVVFQ0JNS1EyRnNhV1p2CmNtNXBZVEVXTUJRR0ExVUVCeE1OVTJGdUlFWnlZVzVqYVhOamJ6RU1NQW9HQTFVRUN4TURRMDlRTVI4d0hRWUQKVlFRREV4WndaV1Z5TUM1dmNtY3hMbVY0WVcxd2JHVXVZMjl0TUZrd0V3WUhLb1pJemowQ0FRWUlLb1pJemowRApBUWNEUWdBRVo4UzRWNzFPQkpweU1JVlpkd1lkRlhBY2tJdHJwdlNyQ2YwSFFnNDBXVzlYU29PT083NkkrVW1mCkVrbVRsSUpYUDcvQXlSUlNSVTM4b0k4SXZ0dTRNNk5OTUVzd0RnWURWUjBQQVFIL0JBUURBZ2VBTUF3R0ExVWQKRXdFQi93UUNNQUF3S3dZRFZSMGpCQ1F3SW9BZ2luT1JJaG5QRUZaVWhYbTZlV0JrbTdLN1pjOFI0L3o3TFc0SApvc3NEbENzd0NnWUlLb1pJemowRUF3SURSd0F3UkFJZ1Zpa0lVWnpnZnVGc0dMUUhXSlVWSkNVN3BEYUVUa2F6ClB6RmdzQ2lMeFVBQ0lDZ3pKWWxXN252WnhQN2I2dGJldTN0OG1yaE1YUXM5NTZtRDQrQm9LdU5JCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0KEhg6Rz4AFR7yI+3X1bdhVk7rPS+piryTqagShwMKhAMKgQMIARIREg9oZWxsb3dvcmxkX3Rlc3Qa6QIKCF9faW52b2tlCtwCQ29BQ0diRTRiRnBRYXViekF3L1hRUFJKbnVEMmJDM1pjQ1Y0YjlUQzBnbXJQSWUrWUJZTUhZYjVXRGhOQVByZ0p2U0NMaUpiWGIxNXF6cDFDMEkwbGhTVmFkVkxINDY1Q0w1aHVMYkR3dG1Mc3dwR0RWbHN3SFBIbWlrK2NlN1h4MDRoRXFOdXB1M1ZVeGFyNmRVZGNVaDh3eXlXMmZnai9xMWY5NlpBbTRTSC9EQWpkdTd5M3FTcUlsUHAxTE1weis3U3BNNi9BVkFQM2FTdURldG9veVRpdERqWU1IVTdPVXhrajJwSDQxTVNxQTBnNjUvYnN6UFFYS2JjYzQ5Rml2d25LRUt5SVY4S3lYVnJQQjFzOEprcFA4OUhmQ1F4bWtSVTNwbXVrTWk0amtMeFlUNHRYejRNQ3diSkpwNEs2YWxlQ2F6MXNqZEhPTmNJQnZlRmIvbmVrUT09EkYwRAIgUfrI86tO/OpDNEny4GHP4N3p8HWza0rEpZwPZfB9hEYCIH8P7UCwBQE9gQ8Q6k6PfKMC8L2r6ayNIT61KyZK4JRyIiAMstDJrHzmQk2UsMjsABWaqNt8Sbiaau1Cnhs+j2pMjipARUI2MDkzRkJFMTgxOTQxNEZFRTczNTE4MDJGNjI2QTk2Njc5MUE2RDg5NDMyQjBFMkI1REY0MEU5M0QxQjRDRhJHMEUCIQDuN5LKLukJ2mCUA8yWBqVO18g2+uDAMniVIrcgOskb8AIgGST4CSedNDJNZXYKkdPrgPb+H8hV+RD7pykLpUBnCSU="]}
```

Response from the transaction is:
```
asset1:100
```

Yay !  You did it !

### Interactive testing

If you want to interactively test FPC, you can use the commands in `../fabric/bin/`. It contains the standard cli commands you expect from fabric -- note: it is though important that you use these scripts rather than use the fabric commands directly. The provide the same interface as the fabric commands but do some additional magic under the cover.   Additionally, there are also two convenience functions for quickly setting up and shutting down a ledger: `ledger_init.sh` and `ledger_shutdown.sh`.
