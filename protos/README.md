# FPC protos

Just run `make`. This will generate FPC protos for C in `$FPC_PATH/common/protos` and for Go in `$FPC_PATH/internal/protos`
To clean and remove all generated files just run `make clean`.

Note that the generated C proto files are required to build the FPC chaincode enclave (`ecc_enclave`).
Special care is needed for the go protos. In order to allow the protos to be used by the FPC client SDK, the generated go protos must be under version control.
In order to prevent spamming FPC commits with generated proto changes, both, `$FPC_PATH/common/protos` and `$FPC_PATH/internal/protos` are in the `.gitignore`.
However, when protos are updates, the resulting go protos must be checked in using `git add -f $FPC_PATH/internal/protos`.
In the future, this process may be automated via CI build, similar to `https://github.com/hyperledger/fabric-protos`.

