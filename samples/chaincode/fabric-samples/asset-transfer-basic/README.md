# Fabric samples exercise

In this exercise we are going to port of the existing Fabric samples chaincode, in particular, the Asset Transfer Basic Chaincode, to a FPC chaincode.

TODO

We are in `$FPC_PATH/samples/chaincode/fabric-samples/asset-transfer/basic`.

## Asset Transfer Basic

### Modify

https://github.com/hyperledger/fabric-samples/tree/main/asset-transfer-basic/chaincode-go

```go
func CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error
```

The only thing we need is to create our own `main.go`.

### Build 

Copy `main.go_template` to `main.go` and fill it in as described in [samples/demos/hlgf22](../../../demos/hlgf22/README.md#write-the-chaincode). Then, with only Go and Docker (also on a Mac):
```bash
make
make docker
```
Inside the dev container (`FPC_CGO=1`), `make` alone builds with ego and packages the SGX image. Without SGX there is no confidentiality; use that for development only.

### Deployment

TODO

### Interacting with the chaincode

```bash
export FABRIC_LOGGING_SPEC=warning
./fpcclient invoke createAsset 101 green 23 Marcus 9999 
./fpcclient query readAsset 101
```
