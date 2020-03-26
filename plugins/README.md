<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->

# Fabric Plugins

The current code base uses custom validation plugins for ecc (`ecc-vscc/`) and ercc (`ercc-vscc/`) to validate transactions. Additionally, ercc uses a decorator plugin (`ercc-decorator/`) to read attestation related data from the peer local filesystem during registration invocation.

## Build plugins

Go plugins are nasty little beasts and often cause pain when building. That is, a go plugin must be build with the same dependencies as used by the application that loads the plugin at runtime. Currently, Fabric still vendors dependencies which make building plugins even harder. So building plugins from the fabric project folder is the easiest way to overcome this at the moment.

We provide a make script that is "smart" enough to build the plugins from within fabric project. Just run 
```
$ cd plugins
$ make
```

Note that `ercc-vscc` plugin will only successfully build when you run `cd $FPC_PATH/ercc; make` before. 

TODO: `ecc-vscc` and `ercc-vscc` plugins are built but cannot be loaded by the peer because of a `golang.org/x/net/` dependency issue.
 