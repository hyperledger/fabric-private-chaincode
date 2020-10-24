package enclave

import (
	"encoding/json"

	"github.com/hyperledger-labs/fabric-private-chaincode/internal/utils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
)

type MockEnclave struct{}

var logger = flogging.MustGetLogger("enclave")

func (MockEnclave) GetRemoteAttestationReport(spid []byte, sig_rl []byte, sig_rl_size uint) ([]byte, []byte, error) {
	panic("implement me")
}

func (MockEnclave) GetLocalAttestationReport(targetInfo []byte) ([]byte, []byte, error) {
	panic("implement me")
}

func (MockEnclave) Invoke(args []byte, pk []byte, shimStub shim.ChaincodeStubInterface) ([]byte, []byte, error) {

	params := &utils.ChaincodeParams{}
	err := json.Unmarshal(args, params)
	if err != nil {
		return nil, nil, err
	}

	logger.Debugf("received args %s", params)

	return []byte("some response"), []byte("enclave signature"), nil
}

func (MockEnclave) GetPublicKey() ([]byte, error) {
	panic("implement me")
}

func (MockEnclave) Create(enclaveLibFile string) error {
	panic("implement me")
}

func (MockEnclave) GetTargetInfo() ([]byte, error) {
	panic("implement me")
}

func (MockEnclave) Bind(report, pk []byte) error {
	panic("implement me")
}

func (MockEnclave) Destroy() error {
	panic("implement me")
}

func (MockEnclave) MrEnclave() (string, error) {
	panic("implement me")
}
