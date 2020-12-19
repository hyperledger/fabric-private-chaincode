<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->

# Build Fabric with FPC support

Run the following command:

    $ cd $FPC_PATH/fabric
    $ make

## Common errors

### Wrong Fabric version
```
Patching Fabric ...
Aborting! Tag on current HEAD () does not match expected tag/v2.3.0!
...
```

Seems that your Fabric is on the wrong branch.
Try to run `pushd $FABRIC_PATH; git checkout tags/v2.3.0; popd;` followed by `make` again.
