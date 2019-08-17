### Tutorial 
This tutorial will show you how to create, build, install and test a chaincode using Fabric-Private-chaincode repository using the build and test framework.

### Prequisites
The repository in https://github.com/hyperledger-labs/fabric-private-chaincode has been installed as per [README.md](https://github.com/hyperledger-labs/fabric-private-chaincode/blob/master/README.md) in FPC-INSTALL-DIR. 

### Create chaincode
* Create a folder named `helloworld`  in FPC-INSTALL-DIR/examples.
* Create 2 files, `helloworld_cc.h` and `helloworld_cc.cpp` as follows.  

File:  helloworld_cc.h
```
#pragma once

#include <string>
std::string putData(std::string asset_name, int value, void*ctx);
std::string marshall (std::string asset_name, uint8_t value);
std::string getData(std::string asset_name, void*ctx);
int invoke(const char* args,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx);

```


File:  helloworld_cc.cpp

```
#include "helloworld_cc.h"
#include "shim.h"
#include <stdbool.h>
#include <stdint.h>
#include "parson.h"
#include "logging.h"

#define OK "OK"
#define NOT_FOUND "Asset not found"

#define MAX_VALUE_SIZE 1024

// implements chaincode logic for invoke
int invoke(const char* args,
    uint8_t* response,
    uint32_t max_response_len,
    uint32_t* actual_response_len,
    void* ctx)
{
    LOG_DEBUG("HelloworldCC: +++ Executing helloworld chaincode invocation +++");
    LOG_DEBUG("HelloworldCC: \tArgs: %s", args);

    std::vector<std::string> argss;
    // parse json args
    unmarshal_args(argss, args);

    std::string function_name = argss[0];
    std::string asset_name = argss[1];
    std::string result;

    if (function_name == "putData")
    {
        int value = std::stoi (argss[2]);
        result = putData(asset_name, value, ctx);
    }
    else if (function_name == "getData")
    {
        result = getData(asset_name, ctx);
    }
    else
    {
        // unknown function
        LOG_DEBUG("HelloworldCC: RECEIVED UNKOWN transaction");
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

//  Add asset_name, value to ledger
std::string putData(std::string asset_name, int value, void*ctx)
{
    LOG_DEBUG("HelloworldCC: +++ putData +++");

    put_state(asset_name.c_str(), (uint8_t*)&value, sizeof(int), ctx);

    return OK;
}

//  Get value set for asset_name in ledger 
std::string getData(std::string asset_name, void*ctx)
{
    std::string result;
    LOG_DEBUG("HelloworldCC: +++ getData +++");

    uint32_t asset_bytes_len = 0;
    uint8_t asset_bytes[MAX_VALUE_SIZE];
    get_state(asset_name.c_str(), asset_bytes, sizeof(asset_bytes), &asset_bytes_len, ctx);

    //  check if asset_name exists
    if (asset_bytes_len > 0)
    {
        //  asset exists;  return value
        result = marshall (asset_name, *asset_bytes);
    }
    else
    {
        //  asset does not exist
        result = NOT_FOUND;
    }
    return result;
}

std::string marshall (std::string asset_name, uint8_t value)
{
    JSON_Value* root_value = json_value_init_object();
    JSON_Object* root_object = json_value_get_object(root_value);
    json_object_set_string(root_object, "name", asset_name.c_str());
    json_object_set_number(root_object, "value", value);
    char* serialized_string = json_serialize_to_string(root_value);
    std::string out(serialized_string);
    json_free_serialized_string(serialized_string);
    json_value_free(root_value);
    return out;
}

```
### Build 
To build the chaincode, add `CMakefile.txt` and `Makefile` (Note that CMakeLists.txt is a modified version of examples/auction/CMakeLists.txt.  Makefile is the same as in examples/auction/Makefile).
File: CMakeLists.txt
```
cmake_minimum_required(VERSION 3.5.1)

set(SOURCE_FILES
    helloworld_cc.cpp
    )

include(../../ecc_enclave/enclave/CMakeLists-common-app-enclave.txt)
```


Make sure you have the environment variables set.  Now, execute:


```make```


If you dont see any errors, you are good to go.

Optional: For this `helloworld` chaincode to be included in the build framework, modify the file: `FPC-INSTALL-DIR/examples/Makefile` to include the `helloworld` chaincode.

File: Makefile   
```TOP = ..
include $(TOP)/build.mk

EXAMPLES = auction echo helloworld

build clean:
	$(foreach DIR, $(EXAMPLES), $(MAKE) -C $(DIR) $@ || exit;)
```
  
Next step is to install, instantiate the chaincode and invoke transactions.  For this, 
you use the integration framework.  Move to `FPC-INSTALL-DIR/integration`.   In the existing `Makefile`,  make these modifications: 

```...
test: auction_test echo_test deployment_test helloworld_test
...
helloworld_test:
        ./helloworld_test.sh
```        
### Time to test !  
Create a file to install, instantiate the chaincode, submit transactions.  The following test script invokes the transaction, putData thrice;  after each time, invokes getData to read the value set before. 
In `FPC-INSTALL-DIR/integration` folder, Create file `helloworld_test.sh`
```

SCRIPTDIR="$(dirname $(readlink --canonicalize ${BASH_SOURCE}))"
FPC_TOP_DIR="${SCRIPTDIR}/.."
CONFIG_HOME="${SCRIPTDIR}/config"
FABRIC_SCRIPTDIR="${FPC_TOP_DIR}/fabric/bin/"

. ${FABRIC_SCRIPTDIR}/lib/common_utils.sh
. ${FABRIC_SCRIPTDIR}/lib/common_ledger.sh

CC_ID=helloworld_test

#this is the path that will be used for the docker build of the chaincode enclave
ENCLAVE_SO_PATH=examples/helloworld/_build/lib/

CC_VERS=0
num_rounds=3

helloworld_test() {
    # install, init, and register (auction) chaincode
    try ${PEER_CMD} chaincode install -l fpc-c -n ${CC_ID} -v ${CC_VERS} -p ${ENCLAVE_SO_PATH}
    sleep 3

    try ${PEER_CMD} chaincode instantiate -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -v ${CC_VERS} -c '{"args":["init"]}' -V ecc-vscc
    sleep 3

    try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["setup", "ercc"]}' --waitForEvent

    try ${PEER_CMD} chaincode query -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["getEnclavePk"]}'

    say "do hello world"
    for (( i=1; i<=$num_rounds; i++ ))
    do
        # submit data $i for "asset1"
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"putData\",\"asset1\",\"'$i'\"]", ""]}' --waitForEvent
        
	# get data for "asset1";  should be $i;
        try ${PEER_CMD} chaincode invoke -o ${ORDERER_ADDR} -C ${CHAN_ID} -n ${CC_ID} -c '{"Args":["[\"getData\",\"asset1\"]", ""]}' --waitForEvent
    done
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

Execute:  
```
cd FPC-INSTALL-DIR/integration
make helloworld_test
```

Output of transaction invocations will look like: 
```
...
2019-08-16 20:38:54.261 UTC [chaincodeCmd] chaincodeInvokeOrQuery -> INFO 002 Chaincode invoke successful. result: status:200 payload:"{\"ResponseData\":\"T0s=\",\"Signature\":\"MEUCIQCMBc1xibInjO4Dc4Ti9w6eXetVDc7G6l6c74NuVanO2wIgZiL1LXF0/EuUZi8y9osXq/1dIvPO1AEgh4fIThB3pX4=\",\"PublicKey\":\"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfKNgCOQLI7lawzyWInbdXzyT+qE889vgyf7bFrx0yVZ+XlyOIrMRgFIxvpeUHxystUcJbfPlx85fv6zfsSybkA==\"}" 
2019-08-16 20:38:56.539 UTC [chaincodeCmd] ClientWait -> INFO 001 txid [6ef1b31334ae727ab15da09f485c6c6455800a82397863fc955259bfabbb753e] committed with status (VALID) at 
2019-08-16 20:38:56.539 UTC [chaincodeCmd] chaincodeInvokeOrQuery -> INFO 002 Chaincode invoke successful. result: status:200 payload:"{\"ResponseData\":\"eyJuYW1lIjoiYXNzZXQxIiwidmFsdWUiOjN9\",\"Signature\":\"MEYCIQDzHPRaPULt/udhxseFZca3Pu/B4cszfgdW6gHvGXhXUAIhAPMxragYhCFqnJWdNzW8l4vxCVlpzeGjyfpppR98OlW/\",\"PublicKey\":\"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEfKNgCOQLI7lawzyWInbdXzyT+qE889vgyf7bFrx0yVZ+XlyOIrMRgFIxvpeUHxystUcJbfPlx85fv6zfsSybkA==\"}" 
helloworld_test.sh: - shutdown ledger


 helloworld_test.sh: Helloworld test PASSED 
 ...
```

Response from the transactions is marked by `Response Data` and it is base64 encoded. 

Yay !  You did it ! 


