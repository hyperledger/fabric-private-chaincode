module github.com/hyperledger-labs/fabric-private-chaincode/ercc

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

require (
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/hyperledger/fabric v1.4.0-rc1.0.20200715015833-3741860ac90f
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200511190512-bcfeb58dd83a
	github.com/hyperledger/fabric-protos-go v0.0.0-20200707132912-fee30f3ccd23
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sykesm/zap-logfmt v0.0.3 // indirect
	golang.org/x/net v0.0.0-20200320220750-118fecf932d8 // indirect
	golang.org/x/sys v0.0.0-20200320181252-af34d8274f85 // indirect
	golang.org/x/tools v0.0.0-20200323164354-18ea2c8f7359 // indirect
	google.golang.org/genproto v0.0.0-20200319113533-08878b785e9c // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
