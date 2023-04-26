module github.com/hyperledger/fabric-private-chaincode

go 1.20

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

// we pin the grpc dep to the version the fabric-sdk-go uses to avoid breaking changes
replace google.golang.org/grpc => google.golang.org/grpc v1.29.1

require (
	github.com/client9/misspell v0.3.4
	github.com/golang/protobuf v1.5.3
	github.com/hyperledger/fabric v1.4.0-rc1.0.20230405174026-695dd57e01c2
	github.com/hyperledger/fabric-chaincode-go v0.0.0-20230228194215-b84622ba6a7a
	github.com/hyperledger/fabric-contract-api-go v1.2.1
	github.com/hyperledger/fabric-protos-go v0.3.0
	github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go v0.0.0-20230505123407-84f9ba1dc4ec
	github.com/hyperledger/fabric-sdk-go v1.0.1-0.20221020141211-7af45cede6af
	github.com/maxbrunsfeld/counterfeiter/v6 v6.6.1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/ginkgo/v2 v2.9.2
	github.com/onsi/gomega v1.27.6
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.7.0
	github.com/stretchr/testify v1.8.2
	golang.org/x/sync v0.1.0
	golang.org/x/tools v0.8.0
	google.golang.org/protobuf v1.30.0
	gopkg.in/yaml.v2 v2.4.0
	honnef.co/go/tools v0.4.3
)

require (
	github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1 // indirect
	github.com/BurntSushi/toml v1.2.1 // indirect
	github.com/IBM/idemix v0.0.0-20220112103229-701e7610d405 // indirect
	github.com/IBM/mathlib v0.0.0-20220112091634-0a7378db6912 // indirect
	github.com/Knetic/govaluate v3.0.1-0.20171022003610-9aa49832a739+incompatible // indirect
	github.com/Microsoft/go-winio v0.6.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/cloudflare/cfssl v1.4.1 // indirect
	github.com/consensys/gnark-crypto v0.6.0 // indirect
	github.com/containerd/containerd v1.6.18 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/docker v23.0.2+incompatible // indirect
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.5.0 // indirect
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/fsouza/go-dockerclient v1.9.7 // indirect
	github.com/go-kit/kit v0.10.0 // indirect
	github.com/go-logfmt/logfmt v0.5.1 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.8 // indirect
	github.com/go-openapi/swag v0.21.1 // indirect
	github.com/go-task/slim-sprig v0.0.0-20230315185526-52ccab3ef572 // indirect
	github.com/gobuffalo/envy v1.10.1 // indirect
	github.com/gobuffalo/packd v1.0.1 // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.4.4 // indirect
	github.com/google/certificate-transparency-go v1.0.21 // indirect
	github.com/google/go-cmp v0.5.9 // indirect
	github.com/google/pprof v0.0.0-20210407192527-94a9f03dee38 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware v1.4.0 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hyperledger/fabric-amcl v0.0.0-20221107192335-5c75bc7be9c0 // indirect
	github.com/hyperledger/fabric-config v0.1.0 // indirect
	github.com/hyperledger/fabric-lib-go v1.0.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/joho/godotenv v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/klauspost/compress v1.15.14 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/miekg/pkcs11 v1.1.1 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/mmcloughlin/addchain v0.4.0 // indirect
	github.com/moby/patternmatcher v0.5.0 // indirect
	github.com/moby/sys/sequential v0.5.0 // indirect
	github.com/moby/term v0.0.0-20210619224110-3f7ff695adc6 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/nxadm/tail v1.4.8 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20211202183452-c5a74bcca799 // indirect
	github.com/opencontainers/runc v1.1.5 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.32.1 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/spf13/viper v1.7.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/sykesm/zap-logfmt v0.0.4 // indirect
	github.com/weppos/publicsuffix-go v0.5.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20190905194746-02993c407bfb // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/zmap/zcrypto v0.0.0-20190729165852-9051775e6a2e // indirect
	github.com/zmap/zlint v0.0.0-20190806154020-fd021b4cfbeb // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	go.uber.org/zap v1.24.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/exp/typeparams v0.0.0-20221208152030-732eee02a75a // indirect
	golang.org/x/mod v0.10.0 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/grpc v1.53.0 // indirect
	gopkg.in/ini.v1 v1.51.0 // indirect
	gopkg.in/tomb.v1 v1.0.0-20141024135613-dd632973f1e7 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
