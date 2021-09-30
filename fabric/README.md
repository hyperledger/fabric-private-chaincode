<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->

# Fetch default Fabric binaries

Run the following command:

```bash
cd $FPC_PATH/fabric
make
```

# Build custom Fabric binaries
The following assumes that the `FABRIC_PATH` environment variable contains the path to Fabric's source code.
```bash
cd $FPC_PATH/fabric
make native
```

Note that this applies all fabric code patches residing in `patches`.
This is optional and not required in order to use FPC.
To clean the native build, type `cd $FPC_PATH/fabric; make clean-native`.
## Common errors

### Wrong Fabric version
```
Patching Fabric ...
Aborting! Tag on current HEAD () does not match expected tag/v2.3.3!
...
```

Seems that your Fabric is on the wrong branch.
Try to run `pushd $FABRIC_PATH; git checkout tags/v2.3.3; popd;` followed by `make` again.
