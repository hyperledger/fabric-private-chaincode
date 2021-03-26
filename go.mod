module github.com/hyperledger-labs/fabric-private-chaincode

go 1.15

// Note:
// - fabric has a go.mod but the normal tagging, e.g., v2.2.0 does NOT
//   follow go module versioning, where API with version > v1 have to be explicitly
//   versioned.
//   the workaround is to update the module not based on version tag, e.g.,
//      go get github.com/hyperledger/fabric@v2.2.0
//   (which will fail) but using the commit id or a branch name
//      go get github.com/hyperledger/fabric@release-2.2
//   The version attributed to, though, seems rather random but, oh, well, ....
// - furthermore, try to keep versions here as much as possible in sync
//   and go mod tidy'ed as additional or newer dependencies can pull in
//   versions which make fabric tools, e.g., configtxgen, fail mysteriously
//   at runtime. (Note though keeping them identical in version will often
//   not be possible ....)

require (
	github.com/golang/protobuf v1.4.2
	github.com/hyperledger/fabric v1.4.0-rc1.0.20201118191903-ec81f3e74fa1
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20201119163726-f8ef75b17719
	github.com/hyperledger/fabric-contract-api-go v1.1.1
	github.com/hyperledger/fabric-protos-go v0.0.0-20201028172056-a3136dde2354
	github.com/hyperledger/fabric-samples/chaincode/marbles02/go v0.0.0-20201218135726-377ccb0b59d5 // indirect
	github.com/hyperledger/fabric-sdk-go v1.0.0-rc1
	github.com/maxbrunsfeld/counterfeiter/v6 v6.3.0
	github.com/onsi/ginkgo v1.12.1
	github.com/onsi/gomega v1.10.3
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	golang.org/x/tools v0.0.0-20201023174141-c8cfbd0f21e6
	google.golang.org/protobuf v1.25.0
)
