<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->

# Build Fabric with FPC support

To build the Fabric peer manually make sure you have built [tlcc_enclave](../tlcc_enclave) before.
Then run the following command:

    $ cd $FPC_PATH/fabric
    $ make

## Common errors

### TLCC not built yet

```
# github.com/hyperledger/fabric/cmd/peer
/usr/local/go/pkg/tool/linux_amd64/link: running gcc failed: exit status 1
/usr/bin/ld: cannot find -ltl
collect2: error: ld returned 1 exit status
```

Seems that the Trusted Ledger has not been built before.
Try to run `pushd $FPC_PATH/tlcc_enclave; make; popd;` followed by `make` again.


### Wrong Fabric version
```
Patching Fabric ...
Aborting! Tag on current HEAD () does not match expected tag/v2.1.1!
...
```

Seems that your Fabric is on the wrong branch.
Try to run `pushd $FABRIC_PATH; git checkout tags/v2.1.1; popd;` followed by `make` again.
