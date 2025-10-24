module github.com/hyperledger/fabric-private-chaincode/samples/chaincode/confidential-escrow

go 1.24.2

require (
	github.com/hyperledger-labs/cc-tools v1.0.2
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20230228194215-b84622ba6a7a
	github.com/hyperledger/fabric-private-chaincode v0.0.0-00010101000000-000000000000
	github.com/hyperledger/fabric-protos-go v0.3.0
)

require (
	github.com/Shopify/sarama v0.0.0-00010101000000-000000000000 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.3.1 // indirect
	github.com/hashicorp/go-version v1.7.0 // indirect
	github.com/hyperledger/fabric v2.1.1+incompatible // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.25.0 // indirect
	golang.org/x/crypto v0.38.0 // indirect
	golang.org/x/net v0.40.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20231016165738-49dd2c1f3d0b // indirect
	google.golang.org/grpc v1.59.0 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
)

replace github.com/hyperledger/fabric-private-chaincode => ../../../

replace github.com/Shopify/sarama => github.com/IBM/sarama v1.45.2
