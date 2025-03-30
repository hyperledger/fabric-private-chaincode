package test_test

import (
	"os"
	"testing"

	fpc "github.com/hyperledger/fabric-private-chaincode/client_sdk/go/pkg/gateway"
	"github.com/hyperledger/fabric-private-chaincode/integration/client_sdk/go/utils"
	"github.com/stretchr/testify/assert"
)

func TestMustNotCrash(t *testing.T) {
	ccID := os.Getenv("CC_ID")
	channelID := os.Getenv("CHAN_ID")
	assert.NotEmpty(t, ccID)
	assert.NotEmpty(t, channelID)
	t.Logf("Use channel: %v, chaincode ID: %v", channelID, ccID)

	network, err := utils.SetupNetwork(channelID)
	assert.NoError(t, err)
	contract := fpc.GetContract(network, ccID)

	// this bidder name might cause the enclave crashing due to a null dereferencing
	// https://github.com/hyperledger/fabric-private-chaincode/blob/88b7c21cc398ed7273d807976f6dffe5e69bd18c/ecc_enclave/enclave/shim.cpp#L195
	result, err := contract.SubmitTransaction("store", "auction", "baz\",\"value\":4141},{\"key\":\"aa", "200")
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(result))

	result, err = contract.EvaluateTransaction("retrieve", "auction")
	assert.NoError(t, err)
	assert.Equal(t, "STILL_ALIVE", string(result))

	result, err = contract.SubmitTransaction("store", "auction", "john", "200")
	assert.NoError(t, err)
	assert.Equal(t, "OK", string(result))

	result, err = contract.EvaluateTransaction("retrieve", "auction")
	assert.NoError(t, err)
	assert.Equal(t, "STILL_ALIVE", string(result))
}
