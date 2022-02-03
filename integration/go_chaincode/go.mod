module github.com/hyperledger/fabric-private-chaincode/integration/go_chaincode

go 1.16

replace (
	github.com/fsouza/go-dockerclient => github.com/fsouza/go-dockerclient v1.4.1
	github.com/go-kit/kit => github.com/go-kit/kit v0.7.0
	github.com/hyperledger/fabric => github.com/hyperledger/fabric v1.4.0-rc1.0.20210722174351-9815a7a8f0f7
	github.com/hyperledger/fabric-protos-go => github.com/hyperledger/fabric-protos-go v0.0.0-20201028172056-a3136dde2354
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20181228115726-23731bf9ba55
)

require (
	github.com/hyperledger-labs/fabric-smart-client v0.0.0-20220203111910-89ea48cecb50
	github.com/hyperledger/fabric v1.4.0-rc1.0.20210722174351-9815a7a8f0f7
	github.com/libp2p/go-libp2p-core v0.3.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.1-0.20210116013205-6990a05d54c2
)
