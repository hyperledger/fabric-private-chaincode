module the-simple-testing-network

go 1.16

replace (
	github.com/fsouza/go-dockerclient => github.com/fsouza/go-dockerclient v1.4.1
	github.com/go-kit/kit => github.com/go-kit/kit v0.7.0
	github.com/hyperledger/fabric => github.com/hyperledger/fabric v1.4.0-rc1.0.20210722174351-9815a7a8f0f7
	github.com/hyperledger/fabric-protos-go => github.com/hyperledger/fabric-protos-go v0.0.0-20201028172056-a3136dde2354
	go.etcd.io/etcd => go.etcd.io/etcd v0.5.0-alpha.5.0.20181228115726-23731bf9ba55
)

//replace github.com/hyperledger-labs/fabric-smart-client => ../../../../../../hyperledger-labs/fabric-smart-client

require (
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/hyperledger-labs/fabric-smart-client v0.0.0-20220715073007-5b78fcf2a13d
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
)
