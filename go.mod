module github.com/hyperledger-labs/fabric-private-chaincode

go 1.14

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
	github.com/dustin/go-broadcast v0.0.0-20171205050544-f664265f5a66
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.1
	github.com/golang/protobuf v1.4.1
	github.com/hyperledger/fabric v1.4.0-rc1.0.20200715015833-3741860ac90f
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200511190512-bcfeb58dd83a
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/hyperledger/fabric-samples/chaincode/marbles02/go v0.0.0-20200723181750-8c32a85f6617 // indirect
	github.com/onsi/ginkgo v1.8.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v0.0.0-20150908122457-1967d93db724
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/tools v0.0.0-20200323164354-18ea2c8f7359
	google.golang.org/protobuf v1.25.0 // indirect
)
