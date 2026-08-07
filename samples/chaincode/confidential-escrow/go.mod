module github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow

go 1.26.3

require (
	github.com/golang/protobuf v1.5.4
	github.com/hyperledger-labs/cc-tools v1.0.2
	github.com/hyperledger/fabric-chaincode-go/v2 v2.3.0
	github.com/hyperledger/fabric-private-chaincode v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-protos-go-apiv2 v0.3.7
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/hyperledger/fabric v2.1.1+incompatible // indirect
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20230228194215-b84622ba6a7a // indirect
	github.com/hyperledger/fabric-lib-go v1.1.3 // indirect
	github.com/hyperledger/fabric-protos-go v0.3.0 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.48.0 // indirect
	golang.org/x/net v0.50.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260120174246-409b4a993575 // indirect
	google.golang.org/grpc v1.79.3 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace github.com/hyperledger/fabric-private-chaincode => ../../../

replace github.com/Shopify/sarama => github.com/IBM/sarama v1.45.2
