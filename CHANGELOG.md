## v1.0.0-rc3
Wed Jan 26 18:02:27 CET 2022

* [5d0290d](https://github.com/hyperledger/fabric-private-chaincode/commit/5d0290d) Release v1.0-rc3
* [0dbf76f](https://github.com/hyperledger/fabric-private-chaincode/commit/0dbf76f) Update IRB readme
* [d472bca](https://github.com/hyperledger/fabric-private-chaincode/commit/d472bca) Link existing FSC examples
* [36a60d3](https://github.com/hyperledger/fabric-private-chaincode/commit/36a60d3) Add del_state support (#640)
* [5febb77](https://github.com/hyperledger/fabric-private-chaincode/commit/5febb77) Fix memory leak in cert decoding
* [3bfeeb5](https://github.com/hyperledger/fabric-private-chaincode/commit/3bfeeb5) Add missing pb_release calls
* [2234bb4](https://github.com/hyperledger/fabric-private-chaincode/commit/2234bb4) Add stress test for many invocations
* [58e58a5](https://github.com/hyperledger/fabric-private-chaincode/commit/58e58a5) Upgrade parson
* [a949ca9](https://github.com/hyperledger/fabric-private-chaincode/commit/a949ca9) Extend FPC Shim support (#637)
* [5713269](https://github.com/hyperledger/fabric-private-chaincode/commit/5713269) fix sgx device location for dcap machines (#639)
* [0aea62d](https://github.com/hyperledger/fabric-private-chaincode/commit/0aea62d) Demo FSC integration using the IRB demo (#635)
* [d07bc32](https://github.com/hyperledger/fabric-private-chaincode/commit/d07bc32) fixup! Upgrade to Fabric 2.3.3
* [eadb34c](https://github.com/hyperledger/fabric-private-chaincode/commit/eadb34c) Upgrade to Fabric 2.3.3
* [079f357](https://github.com/hyperledger/fabric-private-chaincode/commit/079f357) Upgrade to go 1.16.7
* [22f32d8](https://github.com/hyperledger/fabric-private-chaincode/commit/22f32d8) Updated helloworld readme
* [29041c8](https://github.com/hyperledger/fabric-private-chaincode/commit/29041c8) Fixed the version of fabric-ccenv to 2.3.0.
* [b2e72ab](https://github.com/hyperledger/fabric-private-chaincode/commit/b2e72ab) Post-Release: Set version to main

## v1.0.0-rc2
Thu Sep  9 09:57:18 CEST 2021

* [ef710b2](https://github.com/hyperledger/fabric-private-chaincode/commit/ef710b2) Release v1.0-rc2
* [c0e1dfb](https://github.com/hyperledger/fabric-private-chaincode/commit/c0e1dfb) Remove explicit FPC version from docu
* [a007320](https://github.com/hyperledger/fabric-private-chaincode/commit/a007320) catch all chaincode exceptions and fail gracefully with log (#625)
* [d0f2eec](https://github.com/hyperledger/fabric-private-chaincode/commit/d0f2eec) Unittest for ECC (#618)
* [d564662](https://github.com/hyperledger/fabric-private-chaincode/commit/d564662) Separate the FPC go SDK from the Fabric Client SDK Go (#621)
* [6c13943](https://github.com/hyperledger/fabric-private-chaincode/commit/6c13943) Add FPC go protos (#624)
* [dadbd25](https://github.com/hyperledger/fabric-private-chaincode/commit/dadbd25) Fixing Codecov reporting (#622)
* [61fdcb1](https://github.com/hyperledger/fabric-private-chaincode/commit/61fdcb1) HelloWorld on the sample test-network (#619)
* [22ab0dc](https://github.com/hyperledger/fabric-private-chaincode/commit/22ab0dc) Add missing license
* [75653bc](https://github.com/hyperledger/fabric-private-chaincode/commit/75653bc) Fix indentation
* [7e3f242](https://github.com/hyperledger/fabric-private-chaincode/commit/7e3f242) Fix spelling
* [045c329](https://github.com/hyperledger/fabric-private-chaincode/commit/045c329) Add spellchecking for go files
* [7388ac5](https://github.com/hyperledger/fabric-private-chaincode/commit/7388ac5) Fix staticcheck results
* [0ab8501](https://github.com/hyperledger/fabric-private-chaincode/commit/0ab8501) Add go staticchecks
* [84e26db](https://github.com/hyperledger/fabric-private-chaincode/commit/84e26db) New docker flow including pulling images (#612)
* [eb77fbf](https://github.com/hyperledger/fabric-private-chaincode/commit/eb77fbf) Fix Test-Network Readme to shutdown without errors (#615)
* [7c2a146](https://github.com/hyperledger/fabric-private-chaincode/commit/7c2a146) Update attestation verification with additional quote statuses (#610)
* [f9c5a88](https://github.com/hyperledger/fabric-private-chaincode/commit/f9c5a88) Fix path in attestation conversion test
* [f755cec](https://github.com/hyperledger/fabric-private-chaincode/commit/f755cec) Fix build fail and blockchain explorer changes (#609)
* [950cca3](https://github.com/hyperledger/fabric-private-chaincode/commit/950cca3) Adding blockchain explorer to test network (#602)
* [65c021f](https://github.com/hyperledger/fabric-private-chaincode/commit/65c021f) Create pure go crypto impl (#604)
* [eb2b85f](https://github.com/hyperledger/fabric-private-chaincode/commit/eb2b85f) Improve test-network sample to accept any chaincode (#607)
* [a2f6b05](https://github.com/hyperledger/fabric-private-chaincode/commit/a2f6b05) Enable stress tests (#586)
* [66f33b4](https://github.com/hyperledger/fabric-private-chaincode/commit/66f33b4) Revisit Readme
* [51039f1](https://github.com/hyperledger/fabric-private-chaincode/commit/51039f1) Fixed broken links in main readme and helloworld readme
* [dc81cd5](https://github.com/hyperledger/fabric-private-chaincode/commit/dc81cd5) Refactor use of PDO crypto
* [e2dcd7a](https://github.com/hyperledger/fabric-private-chaincode/commit/e2dcd7a) Add hybrid request encryption
* [f97f8dc](https://github.com/hyperledger/fabric-private-chaincode/commit/f97f8dc) Add stress test harness including larger tx parms
* [7522920](https://github.com/hyperledger/fabric-private-chaincode/commit/7522920) Add the usage of fpcclient to the README of test-network deployment.
* [f260421](https://github.com/hyperledger/fabric-private-chaincode/commit/f260421) Fix k8s deployment README
* [aa4bd01](https://github.com/hyperledger/fabric-private-chaincode/commit/aa4bd01) Fix k8s deployment path setting
* [0c8d7d5](https://github.com/hyperledger/fabric-private-chaincode/commit/0c8d7d5) Fix test-network path in simple-go (#580)
* [35c26a6](https://github.com/hyperledger/fabric-private-chaincode/commit/35c26a6) fix linter folders
* [aa732eb](https://github.com/hyperledger/fabric-private-chaincode/commit/aa732eb) update hl-labs -> hl refs
* [a46b759](https://github.com/hyperledger/fabric-private-chaincode/commit/a46b759) Fix test-network setting and README (#573)
* [ec11b8a](https://github.com/hyperledger/fabric-private-chaincode/commit/ec11b8a) Update CODEOWNERS group (#574)
* [28fe6f4](https://github.com/hyperledger/fabric-private-chaincode/commit/28fe6f4) fixup! Remove some make dependencies
* [4ce1854](https://github.com/hyperledger/fabric-private-chaincode/commit/4ce1854) fixup! Remove some make dependencies
* [5bdfe79](https://github.com/hyperledger/fabric-private-chaincode/commit/5bdfe79) Remove some make dependencies
* [4bf0f50](https://github.com/hyperledger/fabric-private-chaincode/commit/4bf0f50) Updates
* [5bd0ddd](https://github.com/hyperledger/fabric-private-chaincode/commit/5bd0ddd) fixup! Make fabric patch optional
* [f210670](https://github.com/hyperledger/fabric-private-chaincode/commit/f210670) Make fabric patch optional
* [97d2caf](https://github.com/hyperledger/fabric-private-chaincode/commit/97d2caf) fixup! New samples structure
* [a2f9dbf](https://github.com/hyperledger/fabric-private-chaincode/commit/a2f9dbf) fixup! New samples structure
* [dd756cf](https://github.com/hyperledger/fabric-private-chaincode/commit/dd756cf) fixup! New samples structure
* [a27adc1](https://github.com/hyperledger/fabric-private-chaincode/commit/a27adc1) fixup! New samples structure
* [34c83ed](https://github.com/hyperledger/fabric-private-chaincode/commit/34c83ed) New samples structure
* [06f59bf](https://github.com/hyperledger/fabric-private-chaincode/commit/06f59bf) Extending HelloWorld tutorial with FPC Client SDK
* [9674b90](https://github.com/hyperledger/fabric-private-chaincode/commit/9674b90) remove unused docker build dependencies
* [9a37217](https://github.com/hyperledger/fabric-private-chaincode/commit/9a37217) Add k8s-based demo network
* [a50c9e1](https://github.com/hyperledger/fabric-private-chaincode/commit/a50c9e1) upgrade pdo
* [7b02fe8](https://github.com/hyperledger/fabric-private-chaincode/commit/7b02fe8) Update Client SDK example with multi-org support
* [b5a75a1](https://github.com/hyperledger/fabric-private-chaincode/commit/b5a75a1) Redoing corrections to solve #540
* [ab52860](https://github.com/hyperledger/fabric-private-chaincode/commit/ab52860) Update "master" branch references to "main"
* [b073fad](https://github.com/hyperledger/fabric-private-chaincode/commit/b073fad) fix
* [37f7fc1](https://github.com/hyperledger/fabric-private-chaincode/commit/37f7fc1) (breaking commit) add test
* [37de010](https://github.com/hyperledger/fabric-private-chaincode/commit/37de010) enable coverage report in CI
* [0dae557](https://github.com/hyperledger/fabric-private-chaincode/commit/0dae557) More Client SDK unit tests
* [0b14603](https://github.com/hyperledger/fabric-private-chaincode/commit/0b14603) Update integration/test-network/README.md
* [dbfb57d](https://github.com/hyperledger/fabric-private-chaincode/commit/dbfb57d) Update integration/test-network/README.md
* [446e978](https://github.com/hyperledger/fabric-private-chaincode/commit/446e978) Revisit test network tutorial
* [7897b83](https://github.com/hyperledger/fabric-private-chaincode/commit/7897b83) Refactor utils in client sdk integration test
* [9dc3351](https://github.com/hyperledger/fabric-private-chaincode/commit/9dc3351) Fix tutorial logging
* [57df801](https://github.com/hyperledger/fabric-private-chaincode/commit/57df801) Add lifecycleclient unit tests
* [a58a2bd](https://github.com/hyperledger/fabric-private-chaincode/commit/a58a2bd) Update MAINTAINERS.md
* [36085f3](https://github.com/hyperledger/fabric-private-chaincode/commit/36085f3) Rename github workflow files
* [db1be5c](https://github.com/hyperledger/fabric-private-chaincode/commit/db1be5c) Repolinter cleanup
* [5ef0e38](https://github.com/hyperledger/fabric-private-chaincode/commit/5ef0e38) Add changelog
* [8089672](https://github.com/hyperledger/fabric-private-chaincode/commit/8089672) Update README
* [6d506e6](https://github.com/hyperledger/fabric-private-chaincode/commit/6d506e6) Update RFC references
* [ad06f0d](https://github.com/hyperledger/fabric-private-chaincode/commit/ad06f0d) Fix issues #536 related with documentation errors in the initial setu… (#537)
* [ad1b5d1](https://github.com/hyperledger/fabric-private-chaincode/commit/ad1b5d1) Add missing FPC v1.0-rc1 to release notes

## v1.0.0-rc1
6 Feb 2021

* [11cc896](https://github.com/hyperledger/fabric-private-chaincode/commit/11cc896) update references and typos
* [2bddb79](https://github.com/hyperledger/fabric-private-chaincode/commit/2bddb79) Release v1.0-rc1
* [6a4039a](https://github.com/hyperledger/fabric-private-chaincode/commit/6a4039a) Add lifecycle support to FPC client SDK (#522)
* [cca8578](https://github.com/hyperledger/fabric-private-chaincode/commit/cca8578) Enable CI for Ubuntu 18.04 and 20.04 + fix READMEs (#532)
* [c66d8f4](https://github.com/hyperledger/fabric-private-chaincode/commit/c66d8f4) Upgrade to Ubuntu 20.04 LTS (#531)
* [c4afaef](https://github.com/hyperledger/fabric-private-chaincode/commit/c4afaef) Fix test-network inside dev container (#529)
* [a9bf600](https://github.com/hyperledger/fabric-private-chaincode/commit/a9bf600) Export also docker configs to dev container
* [7253c65](https://github.com/hyperledger/fabric-private-chaincode/commit/7253c65) fix bugs in additinoal-docker-pkg logic and make more uniform
* [9367a6b](https://github.com/hyperledger/fabric-private-chaincode/commit/9367a6b) Add integration test using client sdk
* [dfd06e6](https://github.com/hyperledger/fabric-private-chaincode/commit/dfd06e6) Replace Travis with Github Action (#525)
* [6454ff3](https://github.com/hyperledger/fabric-private-chaincode/commit/6454ff3) Fix codeblocks in READMEs
* [1d103d3](https://github.com/hyperledger/fabric-private-chaincode/commit/1d103d3) Fix goimports
* [4fe1e46](https://github.com/hyperledger/fabric-private-chaincode/commit/4fe1e46) fixup! Fix golinter filter
* [4374cf4](https://github.com/hyperledger/fabric-private-chaincode/commit/4374cf4) Fix golinter filter
* [59a3c72](https://github.com/hyperledger/fabric-private-chaincode/commit/59a3c72) Clarified wrong comment on SGX 20.04 support
* [ff7d816](https://github.com/hyperledger/fabric-private-chaincode/commit/ff7d816) Some test-net tweaks
* [6a70ad4](https://github.com/hyperledger/fabric-private-chaincode/commit/6a70ad4) Some refactoring
* [0642172](https://github.com/hyperledger/fabric-private-chaincode/commit/0642172) FPC 1.0 Core team
* [399c8ff](https://github.com/hyperledger/fabric-private-chaincode/commit/399c8ff) Docu update for FPC 1.0
* [1ae121e](https://github.com/hyperledger/fabric-private-chaincode/commit/1ae121e) upgrade mock ecc to new message format; fix test-network mock test
* [d61aef0](https://github.com/hyperledger/fabric-private-chaincode/commit/d61aef0) robust handling of constants; comments; use encrypt_message function; move get-request-from-proposal to proto utils
* [ae16560](https://github.com/hyperledger/fabric-private-chaincode/commit/ae16560) enable chaincode request message consistency check
* [496fbd7](https://github.com/hyperledger/fabric-private-chaincode/commit/496fbd7) enable message encryption
* [55430f7](https://github.com/hyperledger/fabric-private-chaincode/commit/55430f7) Fix (ancient) html syntax bug in github templates
* [f669d16](https://github.com/hyperledger/fabric-private-chaincode/commit/f669d16) Make clang-format work in both ubuntus
* [133b661](https://github.com/hyperledger/fabric-private-chaincode/commit/133b661) Bump to SGX SDK 2.12 and mention also Ubuntu 20.04
* [dd597cb](https://github.com/hyperledger/fabric-private-chaincode/commit/dd597cb) Bump FABRIC_VERSION to 2.3.0
* [002fbeb](https://github.com/hyperledger/fabric-private-chaincode/commit/002fbeb) Make client go sdk work with new response message
* [50a54ea](https://github.com/hyperledger/fabric-private-chaincode/commit/50a54ea) support for digital signatures + integration tests work (again) (#501)
* [e736015](https://github.com/hyperledger/fabric-private-chaincode/commit/e736015) bug fix for test-network
* [94cbbaf](https://github.com/hyperledger/fabric-private-chaincode/commit/94cbbaf) Scripting improvement
* [57f5c98](https://github.com/hyperledger/fabric-private-chaincode/commit/57f5c98) Uniform arg & base64 treatment in ecc/contract
* [4b0864a](https://github.com/hyperledger/fabric-private-chaincode/commit/4b0864a) Peer cli assist for client request/response processing
* [c5a0db0](https://github.com/hyperledger/fabric-private-chaincode/commit/c5a0db0) Make initEnclave work again and simplify attestation params
* [3d8ae26](https://github.com/hyperledger/fabric-private-chaincode/commit/3d8ae26) uniform use of FPC_PATH as shell variable
* [1c439a9](https://github.com/hyperledger/fabric-private-chaincode/commit/1c439a9) Test-net Update
* [7ab9b27](https://github.com/hyperledger/fabric-private-chaincode/commit/7ab9b27) ERCC changes
* [163d637](https://github.com/hyperledger/fabric-private-chaincode/commit/163d637) adding enclave id to response (#499)
* [7c1728b](https://github.com/hyperledger/fabric-private-chaincode/commit/7c1728b) Faster docker builds be preloading go downloads
* [fbbea81](https://github.com/hyperledger/fabric-private-chaincode/commit/fbbea81) Support hw attestation params in client sdk
* [347e596](https://github.com/hyperledger/fabric-private-chaincode/commit/347e596) Enforce proper attestation verification in HW mode
* [905e210](https://github.com/hyperledger/fabric-private-chaincode/commit/905e210) Fix docker(-compose) & test-net to work with HW mode
* [8b3148e](https://github.com/hyperledger/fabric-private-chaincode/commit/8b3148e) Cleanup obsolete code
* [2ad4810](https://github.com/hyperledger/fabric-private-chaincode/commit/2ad4810) Transition to request/reply protos inside enclave (#493)
* [66b1e0d](https://github.com/hyperledger/fabric-private-chaincode/commit/66b1e0d) Convert all log to flogger
* [5e9c726](https://github.com/hyperledger/fabric-private-chaincode/commit/5e9c726) Purge more stuff not necessary in FPC 1.0 (aka FPC Lite)
* [8de6f7a](https://github.com/hyperledger/fabric-private-chaincode/commit/8de6f7a) enforce single enclave registation
* [feb9931](https://github.com/hyperledger/fabric-private-chaincode/commit/feb9931) remove static debug encryption key
* [bd296e9](https://github.com/hyperledger/fabric-private-chaincode/commit/bd296e9) state key gen + state encryption
* [1a9a112](https://github.com/hyperledger/fabric-private-chaincode/commit/1a9a112) fixup! Peer Assistance Utility
* [4a77930](https://github.com/hyperledger/fabric-private-chaincode/commit/4a77930) fix mrenclave hex encoding inconsistency
* [dba6c37](https://github.com/hyperledger/fabric-private-chaincode/commit/dba6c37) Peer Assistance Utility
* [89f489a](https://github.com/hyperledger/fabric-private-chaincode/commit/89f489a) chaincode ek message/flow merge (#472)
* [9fa01be](https://github.com/hyperledger/fabric-private-chaincode/commit/9fa01be) Cleanup + bug fix + re-enable (sim-mode only) integration tests (#475)
* [50b5648](https://github.com/hyperledger/fabric-private-chaincode/commit/50b5648) Make fabric-samples work again and easier to use
* [96798bc](https://github.com/hyperledger/fabric-private-chaincode/commit/96798bc) Fix Go refs for non-dockerized environments (#473)
* [a2902d5](https://github.com/hyperledger/fabric-private-chaincode/commit/a2902d5) Bump bl from 1.2.2 to 1.2.3 in /utils/docker-compose/node-sdk
* [eb0322d](https://github.com/hyperledger/fabric-private-chaincode/commit/eb0322d) Bump bl from 1.2.2 to 1.2.3 in /demo/client/backend/fabric-gateway
* [3310106](https://github.com/hyperledger/fabric-private-chaincode/commit/3310106) Bump highlight.js from 9.18.1 to 9.18.5 in /demo/client/frontend
* [8fd9696](https://github.com/hyperledger/fabric-private-chaincode/commit/8fd9696) Godoc for SDK functions (#468)
* [db2a80c](https://github.com/hyperledger/fabric-private-chaincode/commit/db2a80c) Fix license headers
* [fa275ab](https://github.com/hyperledger/fabric-private-chaincode/commit/fa275ab) upgrade to nanopb 0.4.3
* [a26eca9](https://github.com/hyperledger/fabric-private-chaincode/commit/a26eca9) Docker fixes for 20.04
* [6a180c0](https://github.com/hyperledger/fabric-private-chaincode/commit/6a180c0) fixed small protocol inconsistency in diagram
* [57b6bbe](https://github.com/hyperledger/fabric-private-chaincode/commit/57b6bbe) Implement new validation flow (#458)
* [c66ada0](https://github.com/hyperledger/fabric-private-chaincode/commit/c66ada0) Add FPC Lite restrictions to shim.h (#465)
* [4db7275](https://github.com/hyperledger/fabric-private-chaincode/commit/4db7275) createenclave flow
* [ddfb651](https://github.com/hyperledger/fabric-private-chaincode/commit/ddfb651) Update Spec with FPC Lite variant (#457)
* [c6c4186](https://github.com/hyperledger/fabric-private-chaincode/commit/c6c4186) Change flows and protobufs to reflect new enclave discovery
* [812bfa5](https://github.com/hyperledger/fabric-private-chaincode/commit/812bfa5) Match styleguide for fpc protobufs
* [1927cec](https://github.com/hyperledger/fabric-private-chaincode/commit/1927cec) Interface update
* [cb1b4d4](https://github.com/hyperledger/fabric-private-chaincode/commit/cb1b4d4) Remove obsolete ERCC validator and decorators
* [0cd2cf0](https://github.com/hyperledger/fabric-private-chaincode/commit/0cd2cf0) Update sdk test
* [b90d80c](https://github.com/hyperledger/fabric-private-chaincode/commit/b90d80c) Implement attestation conversion in go admin API
* [1cf3d9d](https://github.com/hyperledger/fabric-private-chaincode/commit/1cf3d9d) Make 'make' work ...
* [16c5b81](https://github.com/hyperledger/fabric-private-chaincode/commit/16c5b81) Add initial client go sdk prototype
* [189dac0](https://github.com/hyperledger/fabric-private-chaincode/commit/189dac0) Add fabric-sample/test-network (#451)
* [b8a3ac6](https://github.com/hyperledger/fabric-private-chaincode/commit/b8a3ac6) tmp disable broken integration test and demo
* [5a1eb8b](https://github.com/hyperledger/fabric-private-chaincode/commit/5a1eb8b) ERCC refactoring
* [e913097](https://github.com/hyperledger/fabric-private-chaincode/commit/e913097) enable verify_evidence api in Go
* [6024fdf](https://github.com/hyperledger/fabric-private-chaincode/commit/6024fdf) refactor go log
* [3ccb152](https://github.com/hyperledger/fabric-private-chaincode/commit/3ccb152) move DO_DEBUG down in logging; set logging flags in cmakevariables
* [b5bb701](https://github.com/hyperledger/fabric-private-chaincode/commit/b5bb701) set default logging function to puts
* [f34810b](https://github.com/hyperledger/fabric-private-chaincode/commit/f34810b) make ecc_enclave use logging library, rather than its source code
* [3092ee2](https://github.com/hyperledger/fabric-private-chaincode/commit/3092ee2) add comments on DO_* definitions and compile-time SGX_COMMON_CFLAGS
* [26a46c0](https://github.com/hyperledger/fabric-private-chaincode/commit/26a46c0) port crypto tests to new logging mechanism and re-enabled tests
* [29d0c47](https://github.com/hyperledger/fabric-private-chaincode/commit/29d0c47) turn on new logging mechanism in Go ecc and tlcc
* [c768107](https://github.com/hyperledger/fabric-private-chaincode/commit/c768107) logging as a common library + port ecc, tlcc
* [96f0121](https://github.com/hyperledger/fabric-private-chaincode/commit/96f0121) remove logging.h, switch to common/logging
* [4879345](https://github.com/hyperledger/fabric-private-chaincode/commit/4879345) move logging code under common/logging
* [70e0704](https://github.com/hyperledger/fabric-private-chaincode/commit/70e0704) move edl to common logging
* [ef89d3a](https://github.com/hyperledger/fabric-private-chaincode/commit/ef89d3a) rename to ocall_log
* [77d5d5d](https://github.com/hyperledger/fabric-private-chaincode/commit/77d5d5d) Attestation API  (#444)
* [c3f2508](https://github.com/hyperledger/fabric-private-chaincode/commit/c3f2508) Add new protos as specified with new interfaces
* [3737c27](https://github.com/hyperledger/fabric-private-chaincode/commit/3737c27) Fix broken PUML diagram
* [fcefc5c](https://github.com/hyperledger/fabric-private-chaincode/commit/fcefc5c) Bump http-proxy from 1.18.0 to 1.18.1 in /demo/client/frontend
* [b20a387](https://github.com/hyperledger/fabric-private-chaincode/commit/b20a387) Update UMLs according to the new interfaces
* [7305fca](https://github.com/hyperledger/fabric-private-chaincode/commit/7305fca) Fix broken dev apt add pkg
* [f5e7816](https://github.com/hyperledger/fabric-private-chaincode/commit/f5e7816) Bump nanopb to 0.4.1
* [a42d627](https://github.com/hyperledger/fabric-private-chaincode/commit/a42d627) Build attestation provider.go via build-tags
* [cc9b6b1](https://github.com/hyperledger/fabric-private-chaincode/commit/cc9b6b1) TLCC via go build-tags
* [58c4dca](https://github.com/hyperledger/fabric-private-chaincode/commit/58c4dca) Common attestation api (#433)
* [a59ddd6](https://github.com/hyperledger/fabric-private-chaincode/commit/a59ddd6) TLCC_Enclave<->ECC_Enclave channel interfaces and high-level design
* [175abb6](https://github.com/hyperledger/fabric-private-chaincode/commit/175abb6) whitespace normalization
* [b1366ba](https://github.com/hyperledger/fabric-private-chaincode/commit/b1366ba) New FPC interfaces and state
* [8b7abf2](https://github.com/hyperledger/fabric-private-chaincode/commit/8b7abf2) Current APIs snapshot
* [1cfe977](https://github.com/hyperledger/fabric-private-chaincode/commit/1cfe977) Make epid linkable creds also working with docker
* [f4e3915](https://github.com/hyperledger/fabric-private-chaincode/commit/f4e3915) conditional debug for pdo crypto lib
* [71d9d87](https://github.com/hyperledger/fabric-private-chaincode/commit/71d9d87) update go.* so they hopefully stay invariant without code change
* [555ec99](https://github.com/hyperledger/fabric-private-chaincode/commit/555ec99) Move from obsolete single uae_service lib & header to split version
* [2f1cd40](https://github.com/hyperledger/fabric-private-chaincode/commit/2f1cd40) Make sure switching SGX_MODE work with just a "make clean"
* [01e3e9a](https://github.com/hyperledger/fabric-private-chaincode/commit/01e3e9a) fix bug #416
* [b782058](https://github.com/hyperledger/fabric-private-chaincode/commit/b782058) More quiet docker output to prevent reaching travis log limit termination
* [a4e5908](https://github.com/hyperledger/fabric-private-chaincode/commit/a4e5908) Upgrade to SGX 2.10 and docker image hierachy cleanup, slimdown & speedup
* [9fec7c4](https://github.com/hyperledger/fabric-private-chaincode/commit/9fec7c4) Various minor cleanup
* [86a5302](https://github.com/hyperledger/fabric-private-chaincode/commit/86a5302) Bump elliptic from 6.5.2 to 6.5.3 in /utils/docker-compose/node-sdk
* [17f1a4b](https://github.com/hyperledger/fabric-private-chaincode/commit/17f1a4b) Bump elliptic from 6.5.2 to 6.5.3 in /demo/client/backend/fabric-gateway
* [95e09d6](https://github.com/hyperledger/fabric-private-chaincode/commit/95e09d6) Bump elliptic from 6.5.2 to 6.5.3 in /demo/client/frontend
* [dabbcf4](https://github.com/hyperledger/fabric-private-chaincode/commit/dabbcf4) fix get chaincode definition (#415)
* [642836b](https://github.com/hyperledger/fabric-private-chaincode/commit/642836b) Add ChaincodeDefinition utility functions (#413)
* [1953d93](https://github.com/hyperledger/fabric-private-chaincode/commit/1953d93) Change cli invocations to handle stricter options in 2.2 and more robust peer & ledger scripts
* [af9bd11](https://github.com/hyperledger/fabric-private-chaincode/commit/af9bd11) Build Fabric peer and utilities clean(ish)ly via go modules with no symlink and alike
* [c664066](https://github.com/hyperledger/fabric-private-chaincode/commit/c664066) Upgrade to Fabric v2.2.0
* [51265b2](https://github.com/hyperledger/fabric-private-chaincode/commit/51265b2) Bump lodash from 4.17.15 to 4.17.19 in /demo/client/frontend
* [0beacce](https://github.com/hyperledger/fabric-private-chaincode/commit/0beacce) Bump lodash from 4.17.15 to 4.17.19 in /utils/docker-compose/node-sdk
* [6652dfe](https://github.com/hyperledger/fabric-private-chaincode/commit/6652dfe) Externalization of ERCC's build/launch process + usable crypto lib in ERCC (#396)
* [6dc2801](https://github.com/hyperledger/fabric-private-chaincode/commit/6dc2801) Fix docker-compose fabric-ca image
* [1ed1ddc](https://github.com/hyperledger/fabric-private-chaincode/commit/1ed1ddc) Integrate FPC validation as compiled validator
* [adeb260](https://github.com/hyperledger/fabric-private-chaincode/commit/adeb260) Bump npm-registry-fetch from 8.0.0 to 8.1.1 in /demo/client/frontend

## cr2.0.0
2 Jul 2020

* [adef1ae](https://github.com/hyperledger/fabric-private-chaincode/commit/adef1ae) Version Tech Preview first release to cr2.0.0.0
* [da1a8c8](https://github.com/hyperledger/fabric-private-chaincode/commit/da1a8c8) update Tech Preview to README
* [8aabda4](https://github.com/hyperledger/fabric-private-chaincode/commit/8aabda4) make getRegisteredUsers running again in demo ..
* [1cd5ab1](https://github.com/hyperledger/fabric-private-chaincode/commit/1cd5ab1) no crypto tests on build
* [2550f45](https://github.com/hyperledger/fabric-private-chaincode/commit/2550f45) run pdo tests in build
* [7032371](https://github.com/hyperledger/fabric-private-chaincode/commit/7032371) add pkg-config to docker
* [f2dfb5c](https://github.com/hyperledger/fabric-private-chaincode/commit/f2dfb5c) update to latest pdo
* [3588391](https://github.com/hyperledger/fabric-private-chaincode/commit/3588391) port pdo crypto tests
* [3514f33](https://github.com/hyperledger/fabric-private-chaincode/commit/3514f33) Upgrade docker-compose net to 2.x & re-enable demo
* [7412d7d](https://github.com/hyperledger/fabric-private-chaincode/commit/7412d7d) reference guide plus references to rfc and uml diagrams
* [a7e7384](https://github.com/hyperledger/fabric-private-chaincode/commit/a7e7384) Set MRENCLAVE as version in Chaincode Definition
* [9fc789f](https://github.com/hyperledger/fabric-private-chaincode/commit/9fc789f) Update hello world example
* [fd6c4f2](https://github.com/hyperledger/fabric-private-chaincode/commit/fd6c4f2) UML Diagramm update (#386)
* [24ca548](https://github.com/hyperledger/fabric-private-chaincode/commit/24ca548) createenclave management api (#385)
* [31c43b2](https://github.com/hyperledger/fabric-private-chaincode/commit/31c43b2) Define FPC endorsement policies
* [d950c33](https://github.com/hyperledger/fabric-private-chaincode/commit/d950c33) Adding high-level architecture diagrams (#384)
* [a5cc08d](https://github.com/hyperledger/fabric-private-chaincode/commit/a5cc08d) hide validation and endorsement plugins (#382)
* [c430ccd](https://github.com/hyperledger/fabric-private-chaincode/commit/c430ccd) Update registration diagram (#380)
* [971186f](https://github.com/hyperledger/fabric-private-chaincode/commit/971186f) Logic fix in ercc-vscc
* [c65fd2e](https://github.com/hyperledger/fabric-private-chaincode/commit/c65fd2e) Update key-mgnt diagram (#379)
* [0067512](https://github.com/hyperledger/fabric-private-chaincode/commit/0067512) Revist FPC validation diagram (#378)
* [0c7f643](https://github.com/hyperledger/fabric-private-chaincode/commit/0c7f643) Update invocation and execution diagrams (#377)
* [579785b](https://github.com/hyperledger/fabric-private-chaincode/commit/579785b) Updated key distribution diagram (#375)
* [a5b00eb](https://github.com/hyperledger/fabric-private-chaincode/commit/a5b00eb) FPC Management markdown (#376)
* [9e97b1b](https://github.com/hyperledger/fabric-private-chaincode/commit/9e97b1b) Revisit lifecycle (#374)
* [bd95b2f](https://github.com/hyperledger/fabric-private-chaincode/commit/bd95b2f) component/registration diagrams (#333)
* [479c714](https://github.com/hyperledger/fabric-private-chaincode/commit/479c714) Update FPC lifecycle diagram (#353)
* [fadc9f1](https://github.com/hyperledger/fabric-private-chaincode/commit/fadc9f1) Re-enable ecc-vscc and ercc-vscc
* [bd94291](https://github.com/hyperledger/fabric-private-chaincode/commit/bd94291) Update to Fabric v2.1.1 (including modules!)
* [c5346fc](https://github.com/hyperledger/fabric-private-chaincode/commit/c5346fc) Bump websocket-extensions from 0.1.3 to 0.1.4 in /demo/client/frontend
* [f89bb6b](https://github.com/hyperledger/fabric-private-chaincode/commit/f89bb6b) Make peer & dev builds leaner by excluding unnecessary PDO dependencies
* [66d7cd2](https://github.com/hyperledger/fabric-private-chaincode/commit/66d7cd2) Add nil check to PEM block parsing result
* [235df66](https://github.com/hyperledger/fabric-private-chaincode/commit/235df66) Add missing makefile dependencies to patch & build fabric peer
* [899062c](https://github.com/hyperledger/fabric-private-chaincode/commit/899062c) Replace explicit list of maintainers but group (team)
* [290400f](https://github.com/hyperledger/fabric-private-chaincode/commit/290400f) Build/Test only via dev container and not also directly on host
* [5039ef7](https://github.com/hyperledger/fabric-private-chaincode/commit/5039ef7) Make docker-based tests run also inside dev(elopment) container * make sure host network is also available inside * for volume mounts, make sure the source path maps to a path   understood by docker daemon
* [fcb2d0b](https://github.com/hyperledger/fabric-private-chaincode/commit/fcb2d0b) Make rm failure abort make (Note rm -f does _not_ file if the files do not exist ..)
* [7b0f562](https://github.com/hyperledger/fabric-private-chaincode/commit/7b0f562) Fix path so we can call make also via `make -C utils/docker run`
* [4bb4cb8](https://github.com/hyperledger/fabric-private-chaincode/commit/4bb4cb8) Make missing device error more explicit and make SGX_MODE=SIM default for build - note this is already done for a number of scripts and makes it more fail-safe   from usage-perspective (though not from security :-). Also note that cmake still   has a HW default but that will be removed and referring to the config.mk default   in a separate PR)
* [925e4dd](https://github.com/hyperledger/fabric-private-chaincode/commit/925e4dd) A more robust teardown
* [5a24e7f](https://github.com/hyperledger/fabric-private-chaincode/commit/5a24e7f) More robust and precise handling of docker image cleanup
* [ca7ec72](https://github.com/hyperledger/fabric-private-chaincode/commit/ca7ec72) Version docker images
* [6e43968](https://github.com/hyperledger/fabric-private-chaincode/commit/6e43968) Fix build for concept-release - move both dev and peer to build from local HEAD commit - (try to) minimize re-build of images - add missing dependencies to dev container
* [f671e5a](https://github.com/hyperledger/fabric-private-chaincode/commit/f671e5a) Upgrade to Fabric v2.1.0
* [cd369b8](https://github.com/hyperledger/fabric-private-chaincode/commit/cd369b8) (Hopefully) improve usability of external builder state & log-file (#343)
* [8e3937a](https://github.com/hyperledger/fabric-private-chaincode/commit/8e3937a) exclude pdo from linter and fix indentation (#344)
* [7a2b087](https://github.com/hyperledger/fabric-private-chaincode/commit/7a2b087) Enable linter again (#342)
* [2cb75c2](https://github.com/hyperledger/fabric-private-chaincode/commit/2cb75c2) fix comment in cmake SGX
* [9b4ba80](https://github.com/hyperledger/fabric-private-chaincode/commit/9b4ba80) fix defaults
* [1fd5677](https://github.com/hyperledger/fabric-private-chaincode/commit/1fd5677) fixes
* [52a657c](https://github.com/hyperledger/fabric-private-chaincode/commit/52a657c) temporary use of pdo-crypto lib in ecc_enclave
* [7b13ae1](https://github.com/hyperledger/fabric-private-chaincode/commit/7b13ae1) rename types.h -> fpc-types.h
* [e8b63ee](https://github.com/hyperledger/fabric-private-chaincode/commit/e8b63ee) comments in types.cpp
* [4d66e03](https://github.com/hyperledger/fabric-private-chaincode/commit/4d66e03) more robust SGX_SSL definition
* [f19bd04](https://github.com/hyperledger/fabric-private-chaincode/commit/f19bd04) travis update
* [e9e9fce](https://github.com/hyperledger/fabric-private-chaincode/commit/e9e9fce) fix FPC_PATH definition dependency
* [94777b1](https://github.com/hyperledger/fabric-private-chaincode/commit/94777b1) adding pdo as submodule
* [2feb7d5](https://github.com/hyperledger/fabric-private-chaincode/commit/2feb7d5) compile pdo crypto --from local branch--
* [834ccac](https://github.com/hyperledger/fabric-private-chaincode/commit/834ccac) Let generate_protos.sh immediately fail on error (#340)
* [a8d535e](https://github.com/hyperledger/fabric-private-chaincode/commit/a8d535e) Improved fabric patch/clean - allow for `make clean` in clean state - simplify by removing need for branch - make sure to use tags attached to current commit (rather than "closest" tag) - also included tiny fabic version bugfix in examplse readme
* [3f197a1](https://github.com/hyperledger/fabric-private-chaincode/commit/3f197a1) plantuml versions
* [91d71cd](https://github.com/hyperledger/fabric-private-chaincode/commit/91d71cd) Enabled TLCC for Fabric v2
* [14dcbd2](https://github.com/hyperledger/fabric-private-chaincode/commit/14dcbd2) Upgrade to Protoc 3.11.4
* [02f60f9](https://github.com/hyperledger/fabric-private-chaincode/commit/02f60f9) Update protobufs in tlcc_enclave
* [c49e394](https://github.com/hyperledger/fabric-private-chaincode/commit/c49e394) Move ecc init logic into setup
* [d4d5dde](https://github.com/hyperledger/fabric-private-chaincode/commit/d4d5dde) Complete proxy patch (earlier version contained for unknown reason only part) (#331)
* [adc7e1c](https://github.com/hyperledger/fabric-private-chaincode/commit/adc7e1c) Convert integrations tests to lifecycle v2
* [02b414a](https://github.com/hyperledger/fabric-private-chaincode/commit/02b414a) Remove init chaincode API
* [1068bd2](https://github.com/hyperledger/fabric-private-chaincode/commit/1068bd2) Make peer.sh also work for non-fpc code (and hence also for deployment test)
* [169cfe7](https://github.com/hyperledger/fabric-private-chaincode/commit/169cfe7) tools.go for goimports
* [db76486](https://github.com/hyperledger/fabric-private-chaincode/commit/db76486) Converted integration config files from solo to raft (#324)
* [14a2f58](https://github.com/hyperledger/fabric-private-chaincode/commit/14a2f58) Enable ERCC decorator
* [2294cf3](https://github.com/hyperledger/fabric-private-chaincode/commit/2294cf3) Run FPC CC via External Builders for FPC chaincode - enable external builders in integration test config - modify peer.sh to remove install/instantiate wrapping   for fpc-CC but add a wrapper for lifecycle chaincode package - create the external builder scripts, currently just doing   essentially what the wrapper has done before ..
* [efcde46](https://github.com/hyperledger/fabric-private-chaincode/commit/efcde46) Normalization of Intel's Copyright statement (#321)
* [800365d](https://github.com/hyperledger/fabric-private-chaincode/commit/800365d) fixup! Adapt to recent openssl download reorganization
* [fab86cc](https://github.com/hyperledger/fabric-private-chaincode/commit/fab86cc) Adapt to recent openssl download reorganization
* [04fe9f8](https://github.com/hyperledger/fabric-private-chaincode/commit/04fe9f8) Docker dev run mounts FPC from local fs
* [e08d2cd](https://github.com/hyperledger/fabric-private-chaincode/commit/e08d2cd) Modified intergration/config files to fabric 2.0
* [a945e45](https://github.com/hyperledger/fabric-private-chaincode/commit/a945e45) proxy-patch for fabric v2 peer & re-enable patching in docker & travis
* [f836d87](https://github.com/hyperledger/fabric-private-chaincode/commit/f836d87) Remove broken auction mock cc
* [350aa39](https://github.com/hyperledger/fabric-private-chaincode/commit/350aa39) Remove fabric outdated patches
* [1a82deb](https://github.com/hyperledger/fabric-private-chaincode/commit/1a82deb) Enable Mock server debug logging
* [42b34e4](https://github.com/hyperledger/fabric-private-chaincode/commit/42b34e4) update codeowners
* [eb4b55a](https://github.com/hyperledger/fabric-private-chaincode/commit/eb4b55a) Various minor cleanup - remove go gets from Dockerfile which got obsoleted by go mod - run 'go mod tidy' after build - fixed remaining references to old go shim - added pointer to auction demo description/spec also to top-level demo README.md
* [70f103d](https://github.com/hyperledger/fabric-private-chaincode/commit/70f103d) fix path to fabric bins built from source ...
* [8470b5e](https://github.com/hyperledger/fabric-private-chaincode/commit/8470b5e) init golang modules from existing 'go get' calls
* [c8719d7](https://github.com/hyperledger/fabric-private-chaincode/commit/c8719d7) add PR request template and turn edit instructions into markdown (html) comments
* [42d146b](https://github.com/hyperledger/fabric-private-chaincode/commit/42d146b) fix rocket-chat
* [2c64229](https://github.com/hyperledger/fabric-private-chaincode/commit/2c64229) Templates for github issues
* [498eee8](https://github.com/hyperledger/fabric-private-chaincode/commit/498eee8) move ercc plugins into their own directory
* [d90b2b3](https://github.com/hyperledger/fabric-private-chaincode/commit/d90b2b3) Fix licenses
* [946a39a](https://github.com/hyperledger/fabric-private-chaincode/commit/946a39a) Disable linting for the moment
* [26cfacc](https://github.com/hyperledger/fabric-private-chaincode/commit/26cfacc) Fix includes for golinter
* [9d45641](https://github.com/hyperledger/fabric-private-chaincode/commit/9d45641) Disable integration tests and demo
* [09b2bfc](https://github.com/hyperledger/fabric-private-chaincode/commit/09b2bfc) Disable cc-builder
* [9c4bf7c](https://github.com/hyperledger/fabric-private-chaincode/commit/9c4bf7c) Disable TLCC and plugins (temporary)
* [093d93f](https://github.com/hyperledger/fabric-private-chaincode/commit/093d93f) template package to main
* [2fcbefa](https://github.com/hyperledger/fabric-private-chaincode/commit/2fcbefa) module-ize ercc
* [10e9a6b](https://github.com/hyperledger/fabric-private-chaincode/commit/10e9a6b) Upgrade to go 1.14
* [97f22cc](https://github.com/hyperledger/fabric-private-chaincode/commit/97f22cc) Upgrade to new Fabric v2 chaincode-go package
* [86269a4](https://github.com/hyperledger/fabric-private-chaincode/commit/86269a4) Upgrade to Fabric v2 in docker images and docu
* [1b89c7f](https://github.com/hyperledger/fabric-private-chaincode/commit/1b89c7f) add license and copyright headers to .puml files
* [945bc48](https://github.com/hyperledger/fabric-private-chaincode/commit/945bc48) reorganize old design docu to make ready for merging new design doc
* [88ed274](https://github.com/hyperledger/fabric-private-chaincode/commit/88ed274) key distribution - draft protocol (#153)
* [edf07d5](https://github.com/hyperledger/fabric-private-chaincode/commit/edf07d5) Add enclave lifecycle UML diagrams
* [82d6006](https://github.com/hyperledger/fabric-private-chaincode/commit/82d6006) make avatar download a bit more robust and graceful
* [284e09b](https://github.com/hyperledger/fabric-private-chaincode/commit/284e09b) disclaimer and explicit list of "official" releases
* [3530071](https://github.com/hyperledger/fabric-private-chaincode/commit/3530071) update CODEOWNERS (#263)
* [5d09377](https://github.com/hyperledger/fabric-private-chaincode/commit/5d09377) Bump node deps
* [cbf90f9](https://github.com/hyperledger/fabric-private-chaincode/commit/cbf90f9) Bump acorn from 6.4.0 to 6.4.1 in /demo/client/frontend
* [ad3ffce](https://github.com/hyperledger/fabric-private-chaincode/commit/ad3ffce) fix README.md
* [2b8a0a6](https://github.com/hyperledger/fabric-private-chaincode/commit/2b8a0a6) fix environment
* [6781c34](https://github.com/hyperledger/fabric-private-chaincode/commit/6781c34) Update rocket chat

## cr1.0.1
7 May 2020

* [c94a648](https://github.com/hyperledger/fabric-private-chaincode/commit/c94a648) Concept Release (1.0) with versioned images & able out of the box to run demo (#339)
* [ee471b6](https://github.com/hyperledger/fabric-private-chaincode/commit/ee471b6) Docker related changes * enable containers (and related scripts) to be optionally SGX_MODE aware and run in HW mode iff SGX_MODE=HW * also some dockerfile cleanup
* [13a2097](https://github.com/hyperledger/fabric-private-chaincode/commit/13a2097) (Mostly) Documentation improvements * document options to run network with couchdb and/or hl explorer * document how to run docker-compose behind proxy and improve docu related to proxies * some minimal extensions of the demo docu and a reference thereof in the top README * a few improvements in network setup script
* [1dde82f](https://github.com/hyperledger/fabric-private-chaincode/commit/1dde82f) Updated README to be clearer on how to develop using the Docker Conta… (#237)
* [a5448c9](https://github.com/hyperledger/fabric-private-chaincode/commit/a5448c9) Display published Results correctly
* [76019b2](https://github.com/hyperledger/fabric-private-chaincode/commit/76019b2) Make avatars offline available
* [7235499](https://github.com/hyperledger/fabric-private-chaincode/commit/7235499) Increase progress bar delay
* [39a0d86](https://github.com/hyperledger/fabric-private-chaincode/commit/39a0d86) * package upgrade due to `npm audit fix` ...
* [087ad06](https://github.com/hyperledger/fabric-private-chaincode/commit/087ad06) * graceful reset without logout on mockserver reset (logout can optionally   be enabled based on VUE_APP_LOGOUT_ON_RESET variable)
* [b8d2046](https://github.com/hyperledger/fabric-private-chaincode/commit/b8d2046) * small UI fixes and embellishments ..
* [ff16092](https://github.com/hyperledger/fabric-private-chaincode/commit/ff16092) bug fix: make price point between 0-100, rather than 0-1
* [4956e73](https://github.com/hyperledger/fabric-private-chaincode/commit/4956e73) script modification (auctioneer ends final round manually + manual bid 3 of C-Mobile matches UI except qty of ter 1
* [f6b4420](https://github.com/hyperledger/fabric-private-chaincode/commit/f6b4420) Add Fabric Private Chaincode Wiki Page to README
* [cacafff](https://github.com/hyperledger/fabric-private-chaincode/commit/cacafff) Minor UI updates
* [161fd62](https://github.com/hyperledger/fabric-private-chaincode/commit/161fd62) make mock-server safe for concurrent requests
* [253b155](https://github.com/hyperledger/fabric-private-chaincode/commit/253b155) Editing bid price and quantity (#226)
* [a1b87c8](https://github.com/hyperledger/fabric-private-chaincode/commit/a1b87c8) * run explorer if USE_EXPLORER env var is set to true. Explorer is accessible on port 8090.
* [43b0625](https://github.com/hyperledger/fabric-private-chaincode/commit/43b0625) * run peer with couchdb if USE_COUCHDB env var is set to true. CouchDB viewer is accessible under port 5984.
* [5dd38df](https://github.com/hyperledger/fabric-private-chaincode/commit/5dd38df) * bug-fix in building demo fpc via cc-builder and changing scenario-run so it runs from fresh (unbuilt) repo
* [5dd1c62](https://github.com/hyperledger/fabric-private-chaincode/commit/5dd1c62) Enhance UI
* [38f1c27](https://github.com/hyperledger/fabric-private-chaincode/commit/38f1c27) * add (randomized) delay and mock-reset option to scenario script
* [32d9352](https://github.com/hyperledger/fabric-private-chaincode/commit/32d9352) UI event-based updates auction status
* [7febd30](https://github.com/hyperledger/fabric-private-chaincode/commit/7febd30) Fix Auctione name mapping
* [185a868](https://github.com/hyperledger/fabric-private-chaincode/commit/185a868) Fix auction details
* [064a8da](https://github.com/hyperledger/fabric-private-chaincode/commit/064a8da) Delete chaincode images only if they exist
* [8a8ac7f](https://github.com/hyperledger/fabric-private-chaincode/commit/8a8ac7f) * automate also bootstrapping in start.sh for fabric network setup
* [8112a2a](https://github.com/hyperledger/fabric-private-chaincode/commit/8112a2a) * makefile changes   - add clobber makefile target which cleans also docker images and artifacts     created when running the fabric network setup   - some tweaks to be able to force re-build of docker images   - some additional docker-build targets for auction images as stop-gap until     proxy issue with docker-compose is cleanly solved .. * fix in scenario-run failing with --bootstrap and relative path to script
* [9fae19d](https://github.com/hyperledger/fabric-private-chaincode/commit/9fae19d) Display Assignment Results
* [f356bee](https://github.com/hyperledger/fabric-private-chaincode/commit/f356bee) extend scenario script with publishAssignmentResults
* [a009b5a](https://github.com/hyperledger/fabric-private-chaincode/commit/a009b5a) modify scenario script to avoid control commands for assignment phase
* [aee0bb5](https://github.com/hyperledger/fabric-private-chaincode/commit/aee0bb5) it creates the sym links in build process; these are needed in demo test later
* [897c1e5](https://github.com/hyperledger/fabric-private-chaincode/commit/897c1e5) test depends on build
* [4832160](https://github.com/hyperledger/fabric-private-chaincode/commit/4832160) linter
* [5b9440a](https://github.com/hyperledger/fabric-private-chaincode/commit/5b9440a) fix tests
* [58c7b36](https://github.com/hyperledger/fabric-private-chaincode/commit/58c7b36) add publish assignement results
* [b8f4648](https://github.com/hyperledger/fabric-private-chaincode/commit/b8f4648) add put public state wrappers
* [dbc30ef](https://github.com/hyperledger/fabric-private-chaincode/commit/dbc30ef) check range of impairments
* [7b73193](https://github.com/hyperledger/fabric-private-chaincode/commit/7b73193) add clean target to mock server
* [221d77e](https://github.com/hyperledger/fabric-private-chaincode/commit/221d77e) test.sh creates enclave link in mock server folder
* [56f300d](https://github.com/hyperledger/fabric-private-chaincode/commit/56f300d) short circuit assignment phase, update tests
* [3aca4ce](https://github.com/hyperledger/fabric-private-chaincode/commit/3aca4ce) run chaincode tests in build process
* [a46c925](https://github.com/hyperledger/fabric-private-chaincode/commit/a46c925) extend checks to content lengths returned by ocalls
* [5ed024e](https://github.com/hyperledger/fabric-private-chaincode/commit/5ed024e) fix bug: binary array used as string
* [8230621](https://github.com/hyperledger/fabric-private-chaincode/commit/8230621) * allow re-mapping of MSPId, DN and org for mock server for both requesters   as well as bidders in createAuction requests and getDefaultAuction.
* [f3fd82d](https://github.com/hyperledger/fabric-private-chaincode/commit/f3fd82d) Bring up FPC Auction CC in demo
* [d3ccf3d](https://github.com/hyperledger/fabric-private-chaincode/commit/d3ccf3d) * fix issue #212 and run scenario script in target test so travis will catch bugs in demo ..
* [da41908](https://github.com/hyperledger/fabric-private-chaincode/commit/da41908) * enhanced demo scripting which   - provides in DSL a new verb submit_manual which provides the template     of what should be manually entered   - has dry-run and non-interactive modes for script/json-file debugging and verification   - does extract rc from status code and verifies against expected return code * addressed issues identified in review of PR #210, e.g., added the two previously   deleted territories back and fixed some typos * added missing Makefiles
* [c23afd8](https://github.com/hyperledger/fabric-private-chaincode/commit/c23afd8) Fix ClockBidding in round 1
* [af6a51c](https://github.com/hyperledger/fabric-private-chaincode/commit/af6a51c) TURBO TX COMMIT
* [1ba46c1](https://github.com/hyperledger/fabric-private-chaincode/commit/1ba46c1) Fixing AuctionId param issue and other nids
* [8de71ec](https://github.com/hyperledger/fabric-private-chaincode/commit/8de71ec) Add warning to state viewer
* [9c891d8](https://github.com/hyperledger/fabric-private-chaincode/commit/9c891d8) Revise ui comonents
* [314c9fa](https://github.com/hyperledger/fabric-private-chaincode/commit/314c9fa) Add user avator to menu
* [b442e67](https://github.com/hyperledger/fabric-private-chaincode/commit/b442e67) UI Components cleanup
* [61641ed](https://github.com/hyperledger/fabric-private-chaincode/commit/61641ed) Add demo dashboard including ledger and state views
* [2674a7c](https://github.com/hyperledger/fabric-private-chaincode/commit/2674a7c) Remove broken components
* [4e4a892](https://github.com/hyperledger/fabric-private-chaincode/commit/4e4a892) Upgrade frontend deps
* [f8b0b54](https://github.com/hyperledger/fabric-private-chaincode/commit/f8b0b54) Add Docker Chaincode Builder Script
* [7ccb59c](https://github.com/hyperledger/fabric-private-chaincode/commit/7ccb59c) Signed-off-by: Jeb Linton <jrlinton@us.ibm.com>
* [352fac4](https://github.com/hyperledger/fabric-private-chaincode/commit/352fac4) consolidate mspIds ... (#209)
* [0d3e435](https://github.com/hyperledger/fabric-private-chaincode/commit/0d3e435) bug fix: excess demand vector computation
* [82d3430](https://github.com/hyperledger/fabric-private-chaincode/commit/82d3430) remove crypto dependency, move it to shim.h
* [fc9b79a](https://github.com/hyperledger/fabric-private-chaincode/commit/fc9b79a) fix impairment-adjusted channel price
* [6094ac3](https://github.com/hyperledger/fabric-private-chaincode/commit/6094ac3) fix auction status response (from number to string)
* [cd9f41d](https://github.com/hyperledger/fabric-private-chaincode/commit/cd9f41d) update readme: no submitassign bid implemented
* [0ff24a8](https://github.com/hyperledger/fabric-private-chaincode/commit/0ff24a8) add license to test.sh
* [fdbddd5](https://github.com/hyperledger/fabric-private-chaincode/commit/fdbddd5) spectrum auction chaincode
* [c00743c](https://github.com/hyperledger/fabric-private-chaincode/commit/c00743c) * make tput errors go away for docker execs ...
* [79f04a1](https://github.com/hyperledger/fabric-private-chaincode/commit/79f04a1) Public state and other enhancements (#204)
* [92ef7c7](https://github.com/hyperledger/fabric-private-chaincode/commit/92ef7c7) Demo Scripting ... (#203)
* [4378122](https://github.com/hyperledger/fabric-private-chaincode/commit/4378122) Add end to end scripts to bring up entire demo
* [bb03fb3](https://github.com/hyperledger/fabric-private-chaincode/commit/bb03fb3) Bootstrap script no longer requires arguments
* [0303058](https://github.com/hyperledger/fabric-private-chaincode/commit/0303058) Switch json tag from mspid to mspId
* [3ee2579](https://github.com/hyperledger/fabric-private-chaincode/commit/3ee2579) Copy source files and package-lock.json
* [f93c4b2](https://github.com/hyperledger/fabric-private-chaincode/commit/f93c4b2) Fabric-gateway fixes
* [38ad998](https://github.com/hyperledger/fabric-private-chaincode/commit/38ad998) Use local license script
* [be7a4eb](https://github.com/hyperledger/fabric-private-chaincode/commit/be7a4eb) Add code owners file
* [a7286cb](https://github.com/hyperledger/fabric-private-chaincode/commit/a7286cb) * patch for fabric to mock creator and creation of mocked creator in demo mock backend * use principal in auction template matching names as our registerUsers.sh would create * also make it a principle to spell principal correctly ;-)
* [cc417e6](https://github.com/hyperledger/fabric-private-chaincode/commit/cc417e6) * pass any response objects also on failure case & in client front-end make sure there is always status object
* [4efb5bc](https://github.com/hyperledger/fabric-private-chaincode/commit/4efb5bc) Various bug fixes     * fix issues #195     * mitigate #187 for now by making response buffer conservative 100k     * bug fix to allow FPC chaincode build from arbitrary location
* [b123af4](https://github.com/hyperledger/fabric-private-chaincode/commit/b123af4) * fix issue #177 * make test work also for simulator mode * two other small bug fixes ..
* [15c7367](https://github.com/hyperledger/fabric-private-chaincode/commit/15c7367) Initial clock auction UI (#192)
* [071e377](https://github.com/hyperledger/fabric-private-chaincode/commit/071e377) Add clock auction backend client and folder structure (#191)
* [b7543dd](https://github.com/hyperledger/fabric-private-chaincode/commit/b7543dd) Mock chaincode testing
* [aa1187c](https://github.com/hyperledger/fabric-private-chaincode/commit/aa1187c) Docker-compose fixes and improvements (#193)
* [153fa0e](https://github.com/hyperledger/fabric-private-chaincode/commit/153fa0e) Bump lodash from 4.17.11 to 4.17.15 in /utils/docker-compose/node-sdk
* [f9e6839](https://github.com/hyperledger/fabric-private-chaincode/commit/f9e6839) Bump eslint-utils from 1.3.1 to 1.4.3 in /utils/docker-compose/node-sdk
* [0e1637f](https://github.com/hyperledger/fabric-private-chaincode/commit/0e1637f) Create FPC Docker Compose Network
* [0c7e641](https://github.com/hyperledger/fabric-private-chaincode/commit/0c7e641) Add Plugins Make Target
* [711d5fb](https://github.com/hyperledger/fabric-private-chaincode/commit/711d5fb) Build Base CC Image Only When Needed
* [1b47eae](https://github.com/hyperledger/fabric-private-chaincode/commit/1b47eae) Updated link
* [92242e4](https://github.com/hyperledger/fabric-private-chaincode/commit/92242e4) Add "Contributions Welcome" section
* [98eb021](https://github.com/hyperledger/fabric-private-chaincode/commit/98eb021) - reference also docker dev image in docu and make dev image work for full dev cycle (including integration tests) - make plantuml output type (PLANTUML_IMG_FORMAT) optionaly configurable via make commandline
* [2edf3ce](https://github.com/hyperledger/fabric-private-chaincode/commit/2edf3ce) generate enclave signer's key on the fly
* [6687429](https://github.com/hyperledger/fabric-private-chaincode/commit/6687429) make sure all docker images have consistent name with hyperledger as repo name (#174)
* [274c0e8](https://github.com/hyperledger/fabric-private-chaincode/commit/274c0e8) Add inital travis build file
* [a8e5b96](https://github.com/hyperledger/fabric-private-chaincode/commit/a8e5b96) New docker images
* [9583608](https://github.com/hyperledger/fabric-private-chaincode/commit/9583608) Improved error handling in tlcc and sgxcclib consolidation (#158)
* [1ab1544](https://github.com/hyperledger/fabric-private-chaincode/commit/1ab1544) Update to Fabric v1.4.3
* [5cf58f0](https://github.com/hyperledger/fabric-private-chaincode/commit/5cf58f0) Fixing issue 152
* [214da5c](https://github.com/hyperledger/fabric-private-chaincode/commit/214da5c) Fixing Issue #42
* [15dd559](https://github.com/hyperledger/fabric-private-chaincode/commit/15dd559) - merge with PR #142
* [f332bc1](https://github.com/hyperledger/fabric-private-chaincode/commit/f332bc1) - merge with PR #146
* [344c9ba](https://github.com/hyperledger/fabric-private-chaincode/commit/344c9ba) - fix buggy buffer management in get_state
* [2aebb1b](https://github.com/hyperledger/fabric-private-chaincode/commit/2aebb1b) * change argument handling to be closer to other shims  (WIP)   - replace marshal with GetStringArgs function   - hide json-argument encoding from client (i.e., client passes now args as     {"args": ["arg1", "arg2", ...]} rather than     {"args":["[\"arg1\", \"arg2\", ... ]"]} ...)   - simplify context handling (and strongly type now with shim_ctx_ptr_t     rather than void*)   - part of this also required chanegd enclave transaction processing     which is now a bit more uniform across the different implementations     in particular related to logging. It also required a new tlcc test case     Note, though, there is still an (already existing) signature validation     problem in tlcc, see issue #152, which for now is just made more explicit     by making tlcc_enclave tests verbose (as such errors show up only in log     but not in failing tests)
* [a88dd71](https://github.com/hyperledger/fabric-private-chaincode/commit/a88dd71) Add link to Fabric maintainer presentation
* [807c708](https://github.com/hyperledger/fabric-private-chaincode/commit/807c708) - add implementation of FPC-CC init() and augment auction_test correspondingly
* [d6b126c](https://github.com/hyperledger/fabric-private-chaincode/commit/d6b126c) - more shim.h clean-up and completion   - get_creator_name implementation and "quick fix integration" into auction_test   - dropped chaincode.h   - per-used go & node shim and added a few other (still commented out)     functions which add additional semantics (rather than providing different     ways to do the same, i.e., syntactic sugaring) and which we might consider     to add eventually from completeness & conssitency perspective
* [33dd164](https://github.com/hyperledger/fabric-private-chaincode/commit/33dd164) Added tutorial for first private chaincode (#142)
* [a00b4b8](https://github.com/hyperledger/fabric-private-chaincode/commit/a00b4b8) Integration test fixes (#146)
* [0269606](https://github.com/hyperledger/fabric-private-chaincode/commit/0269606) separate out shim internal definitions from shim.h and add function prototype for function chaincode has to implement
* [b3887d2](https://github.com/hyperledger/fabric-private-chaincode/commit/b3887d2) log-file+line-patch
* [17f8a18](https://github.com/hyperledger/fabric-private-chaincode/commit/17f8a18) Fix broken tlcc config tx parser
* [4cb41ed](https://github.com/hyperledger/fabric-private-chaincode/commit/4cb41ed) "fixed" tlcc_enclave ledger by - generalizing test.c for arbitrary number of blocks - write a getLedgerBlocks scripts to extract ledger blocks - integrated getLedgerBlocks in deployment_test.sh to use the later as a scenario generator for the blocks - above is exposed as target 'gen_testcases' in ./tlcc_enclave - Note: this doesn't really solve issue #143 as the test is not really testing much as in logs i see error but nothing causes the test to fail (not even an out-of-order block!) - also generalized/diversified deployment_test a bit
* [13cab07](https://github.com/hyperledger/fabric-private-chaincode/commit/13cab07) - hide enclave setup functions in wrapper & make separate namespace from internal functions vs functions passed to FPC chaincode - make scripts a bit versatile, also change CONFIG_HOME to (fabric-standard)   FABRIC_CFG_PATH - also disable for now failing tlcc test and punt to issue 139 and 143
* [989feb4](https://github.com/hyperledger/fabric-private-chaincode/commit/989feb4) create dummy ias files in simulator mode if files do not exist
* [2cd5059](https://github.com/hyperledger/fabric-private-chaincode/commit/2cd5059) consolidate go dependency fetch in top-level; current version loads some too late if you start from a vanilla system ...
* [c9a01ee](https://github.com/hyperledger/fabric-private-chaincode/commit/c9a01ee) Add link to contribution tutorial
* [470c3ba](https://github.com/hyperledger/fabric-private-chaincode/commit/470c3ba) Remove obsolete fabric/sgxconfig
* [44cea6c](https://github.com/hyperledger/fabric-private-chaincode/commit/44cea6c) Update documentation reflecting SDKization
* [9910d3d](https://github.com/hyperledger/fabric-private-chaincode/commit/9910d3d) upgrade to fabric v1.4.2 & config upgrade from old release-1.3 version
* [c8bbdd1](https://github.com/hyperledger/fabric-private-chaincode/commit/c8bbdd1) doc update for plantuml dependencies, addressing issue #126
* [bf79337](https://github.com/hyperledger/fabric-private-chaincode/commit/bf79337) Fix concurrency issue in registry (#122)
* [d81f26d](https://github.com/hyperledger/fabric-private-chaincode/commit/d81f26d) Wrapper script and libraries improvements * make peer a bit more robust and more complete * handle also chaincode upgrade & chaincode list * additional utility scripts * fix bug in common_ledger.sh / precond_test to handle case where ledger state directory is used first time during ledger_init * some other minor cleanup
* [0d634b4](https://github.com/hyperledger/fabric-private-chaincode/commit/0d634b4) Updated UML diagrams * extended lifecycle sequence diagram with issues and added programming model/api diagram * change extension of PlantUML file from .txt to .puml
* [e7d504c](https://github.com/hyperledger/fabric-private-chaincode/commit/e7d504c) Enable deployment of multiple private chaincodes (#86) (#108)
* [7829453](https://github.com/hyperledger/fabric-private-chaincode/commit/7829453) ecc docker image built as an actual boilerplate (#109)
* [578b74a](https://github.com/hyperledger/fabric-private-chaincode/commit/578b74a) Better handling of _build folder (#110)
* [8ab6669](https://github.com/hyperledger/fabric-private-chaincode/commit/8ab6669) Separation of shim/enclave from chaincode app (#93)
* [f3c3ba5](https://github.com/hyperledger/fabric-private-chaincode/commit/f3c3ba5) remove double build
* [08164ce](https://github.com/hyperledger/fabric-private-chaincode/commit/08164ce) Load SPID as text (#100)
* [ac5129c](https://github.com/hyperledger/fabric-private-chaincode/commit/ac5129c) * wrap/hide also ercc installation and tlcc join behind channel join * wrap also other fabric commands (although right now they just pass stuff through) * documentation & makefile for plantuml * license headers for docu * some small clean-up
* [04e7f65](https://github.com/hyperledger/fabric-private-chaincode/commit/04e7f65) * add peer wrapper to do proactive docker switcheroo and add first verison uml sequence diagram of fpc lifecyle   (Note: docu is _NOT_ yet update and post-poned to a major overhaul we anyway have to do after the sdk-ization    refactoring)
* [740378e](https://github.com/hyperledger/fabric-private-chaincode/commit/740378e) * program to derive container-name given cc name and version as well as peer id and networkid
* [c3e6fc9](https://github.com/hyperledger/fabric-private-chaincode/commit/c3e6fc9) update licence headers (#95)
* [46c4191](https://github.com/hyperledger/fabric-private-chaincode/commit/46c4191) files formatted with make linter
* [f4e75fc](https://github.com/hyperledger/fabric-private-chaincode/commit/f4e75fc) add echo's to lint
* [766f998](https://github.com/hyperledger/fabric-private-chaincode/commit/766f998) add cpplinter to check format of cpp, c, h files
* [f965982](https://github.com/hyperledger/fabric-private-chaincode/commit/f965982) add clang-format
* [9835a61](https://github.com/hyperledger/fabric-private-chaincode/commit/9835a61) * fix PR #78 where removal of dot-import unhid a bug which got overlooked in testing
* [a5c7465](https://github.com/hyperledger/fabric-private-chaincode/commit/a5c7465) Unit test cleaning (#78)
* [70f3e65](https://github.com/hyperledger/fabric-private-chaincode/commit/70f3e65) Integration Test (#79)
* [fc1b7e1](https://github.com/hyperledger/fabric-private-chaincode/commit/fc1b7e1) * call fabric script directly from fabric instead of a (modified) copy & (#80)
* [720842f](https://github.com/hyperledger/fabric-private-chaincode/commit/720842f) Fix for PR#73 and minor clean up (#75)
* [cba3901](https://github.com/hyperledger/fabric-private-chaincode/commit/cba3901) Support for SGX simulation mode (#57)
* [261e914](https://github.com/hyperledger/fabric-private-chaincode/commit/261e914) Add missing license identifier
* [e30d703](https://github.com/hyperledger/fabric-private-chaincode/commit/e30d703) fix make docker step
* [01e2d92](https://github.com/hyperledger/fabric-private-chaincode/commit/01e2d92) Perform gofmt, goimport, and go vet (#62)
* [aee034c](https://github.com/hyperledger/fabric-private-chaincode/commit/aee034c) Add linter checks to build (#61)
* [8385626](https://github.com/hyperledger/fabric-private-chaincode/commit/8385626) Rename project to Fabric Private Chaincode (#59)
* [73f34e0](https://github.com/hyperledger/fabric-private-chaincode/commit/73f34e0) Adopt Fabric's CoC and contribution guidelines (#60)
* [4737862](https://github.com/hyperledger/fabric-private-chaincode/commit/4737862) One make to rule them all. (#52)
* [c9bb368](https://github.com/hyperledger/fabric-private-chaincode/commit/c9bb368) - convert to new IAS authentication from specification release 5.0 - build and typo fix for selection of linkability patch
* [87fcae8](https://github.com/hyperledger/fabric-private-chaincode/commit/87fcae8) Fixing typos (#46)
* [9ddf14b](https://github.com/hyperledger/fabric-private-chaincode/commit/9ddf14b) user-defined sgx attestation type
* [5ca2ad1](https://github.com/hyperledger/fabric-private-chaincode/commit/5ca2ad1) fix "Public key not valid (Point not on curve)" error
* [64f6ec5](https://github.com/hyperledger/fabric-private-chaincode/commit/64f6ec5) Various streamlining - streamlining of demo (no copy and no editing anymore necessary) - various docu fixes & improvements - moved to fabric v1.4.1 (and tag, not branch) [handles issue #35 completely] - support also for SGX SDK 2.5 [addresses issue #28 as far as support in our code is concerned, doesn't upgrade docker image, though. change is just add import of stdbool.h ..]
* [d446ec4](https://github.com/hyperledger/fabric-private-chaincode/commit/d446ec4) Fix return statements in auction chaincode
* [45b3fff](https://github.com/hyperledger/fabric-private-chaincode/commit/45b3fff) support operations behind proxy (#24)
* [c108e4b](https://github.com/hyperledger/fabric-private-chaincode/commit/c108e4b) run gofmt (#34)
* [697fdb3](https://github.com/hyperledger/fabric-private-chaincode/commit/697fdb3) fix docker-sgx build
* [dda1215](https://github.com/hyperledger/fabric-private-chaincode/commit/dda1215) Update docu (#27)
* [8d94668](https://github.com/hyperledger/fabric-private-chaincode/commit/8d94668) enhance cmake configuration
* [37606f8](https://github.com/hyperledger/fabric-private-chaincode/commit/37606f8) fix lib path for sim mode
* [54ec811](https://github.com/hyperledger/fabric-private-chaincode/commit/54ec811) list of licenses and disclaimer (#25)
* [f089daa](https://github.com/hyperledger/fabric-private-chaincode/commit/f089daa) Update v2 to v3 in ias url
* [858b939](https://github.com/hyperledger/fabric-private-chaincode/commit/858b939) doc on how to make it work with linkable epid credentials
* [2f5f7b9](https://github.com/hyperledger/fabric-private-chaincode/commit/2f5f7b9) docs enhancements for IAS

## fabric_v1.4
5 Mar 2019

* [41c4e6c](https://github.com/hyperledger/fabric-private-chaincode/commit/41c4e6c) update doc
* [5e96aa9](https://github.com/hyperledger/fabric-private-chaincode/commit/5e96aa9) Update to SGX SDK v2.4
* [ffba4d4](https://github.com/hyperledger/fabric-private-chaincode/commit/ffba4d4) update custom validation to fabric 1.4
* [04db993](https://github.com/hyperledger/fabric-private-chaincode/commit/04db993) update nanopb
* [3c35c42](https://github.com/hyperledger/fabric-private-chaincode/commit/3c35c42) update fabric config
* [af35928](https://github.com/hyperledger/fabric-private-chaincode/commit/af35928) Upgrade to Fabric v1.4

## fabric_v1.2
5 Mar 2019

* [d1136c6](https://github.com/hyperledger/fabric-private-chaincode/commit/d1136c6) Update docu
* [7334ef5](https://github.com/hyperledger/fabric-private-chaincode/commit/7334ef5) Enable ercc verification
* [53e0845](https://github.com/hyperledger/fabric-private-chaincode/commit/53e0845) Fix can open ledger
* [d381af7](https://github.com/hyperledger/fabric-private-chaincode/commit/d381af7) fix tlcc doc
* [471b529](https://github.com/hyperledger/fabric-private-chaincode/commit/471b529) fix edl import issue
* [9760c51](https://github.com/hyperledger/fabric-private-chaincode/commit/9760c51) Fix build and improve docu
* [3a49431](https://github.com/hyperledger/fabric-private-chaincode/commit/3a49431) Initial implementation
* [b8e6c88](https://github.com/hyperledger/fabric-private-chaincode/commit/b8e6c88) Initial version
* [0ef197e](https://github.com/hyperledger/fabric-private-chaincode/commit/0ef197e) Initial commit

