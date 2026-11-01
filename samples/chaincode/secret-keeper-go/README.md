# Secret Keeper

Secret Keeper is a demo application designed to securely store sensitive information, acting as a digital vault. It's ideal for users who need to manage access to shared secrets within a team or organization.

## Functions

Secret Keeper provides the following functionalities:

- **InitSecretKeeper**: Initializes the application with default authorization and secret values. Intended for one-time use at application setup. Note: While potential misuse is considered low-risk, it's recommended to secure access to this function.

- **RevealSecret**: Allows authorized users to view the currently stored secret.

- **LockSecret**: Enables authorized users to update the secret value. This action replaces the existing secret.

- **AddUser**: Permits authorized users to add a new user to the authorization list, granting them access to all functions.

- **RemoveUser**: Allows authorized users to remove an existing user from the authorization list, revoking their access.

## Run without SGX

The sample has two mains: `cmd/naive` (the default) stores each key as is, `cmd/skvs` stores all state as one encrypted key (`fpc.WithSKVS()`). With only Go and Docker (also on a Mac):
```bash
make             # compile cmd/naive with plain go
make docker      # build the fpc/secret-keeper-go image
# or, for cmd/skvs:
make docker ECC_MAIN_FILES=cmd/skvs/main.go
```
Then deploy it as described in [samples/deployment/test-network](../../deployment/test-network/README.md#without-sgx-eg-on-a-mac), with `CC_ID=secret-keeper-go` and `CC_NAME=secret-keeper-go`, or with [the-simple-testing-network](../../deployment/fabric-smart-client/the-simple-testing-network) as below.

Without SGX, FPC gives no confidentiality; the chaincode says so in its log at startup. Use this for development only. FPC-go's attestation is always simulated today.

## Example Usage

To demonstrate Secret Keeper's capabilities, you can deploy the chaincode to [the-simple-testing-network](https://github.com/hyperledger/fabric-private-chaincode/tree/main/samples/deployment/fabric-smart-client/the-simple-testing-network) and then invoke it with the [simple-cli-go](https://github.com/hyperledger/fabric-private-chaincode/tree/main/samples/application/simple-cli-go).

1. Initialize Secret Keeper:
```
./fpcclient invoke initSecretKeeper
```
2. Reveal the secret as Alice:
```
./fpcclient query revealSecret Alice
```
3. Change the secret as Bob:
```
./fpcclient invoke lockSecret Bob NewSecret
```
4. Attempt to reveal the secret as Alice (now updated):
```
./fpcclient query revealSecret Alice
```
5. Remove Bob's access as Alice:
```
./fpcclient invoke removeUser Alice Bob
```
6. Attempt to reveal the secret as Bob (should fail):
```
./fpcclient query revealSecret Bob // (will failed) 
```
7. Re-add Bob to the authorization list as Alice:
```
./fpcclient invoke addUser Alice Bob
```
8. Bob can now reveal the secret successfully:
```
./fpcclient query revealSecret Bob // (will success)
```
