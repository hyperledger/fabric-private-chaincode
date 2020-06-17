module github.com/hyperledger-labs/fabric-private-chaincode

go 1.14

require (
	github.com/Knetic/govaluate v3.0.0+incompatible // indirect
	github.com/Shopify/sarama v1.26.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.5.7 // indirect
	github.com/dustin/go-broadcast v0.0.0-20171205050544-f664265f5a66
	github.com/fsouza/go-dockerclient v1.6.5 // indirect
	github.com/gin-contrib/cors v1.3.1
	github.com/gin-gonic/gin v1.6.1
	github.com/golang/protobuf v1.3.5
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hyperledger-labs/fabric-private-chaincode/ercc v0.0.0-20200323171133-70f103dd66ee
	github.com/hyperledger/fabric v2.1.1+incompatible
	github.com/hyperledger/fabric-amcl v0.0.0-20200424173818-327c9e2cf77a // indirect
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20200511190512-bcfeb58dd83a
	github.com/hyperledger/fabric-lib-go v1.0.0 // indirect
	github.com/hyperledger/fabric-protos-go v0.0.0-20200506201313-25f6564b9ac4
	github.com/miekg/pkcs11 v1.0.3 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/pelletier/go-toml v1.6.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.6.2
	github.com/syndtr/goleveldb v1.0.0 // indirect
	github.com/tedsuo/ifrit v0.0.0-20191009134036-9a97d0632f00 // indirect
	github.com/willf/bitset v1.1.10 // indirect
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	golang.org/x/tools v0.0.0-20200323164354-18ea2c8f7359
	gopkg.in/ini.v1 v1.55.0 // indirect
)

replace github.com/hyperledger-labs/fabric-private-chaincode/ercc => ./ercc
