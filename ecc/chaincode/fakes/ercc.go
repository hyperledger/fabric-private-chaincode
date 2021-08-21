// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-private-chaincode/internal/protos"
)

type ErccStub struct {
	QueryEnclaveCredentialsStub        func(shim.ChaincodeStubInterface, string, string, string) (*protos.Credentials, error)
	queryEnclaveCredentialsMutex       sync.RWMutex
	queryEnclaveCredentialsArgsForCall []struct {
		arg1 shim.ChaincodeStubInterface
		arg2 string
		arg3 string
		arg4 string
	}
	queryEnclaveCredentialsReturns struct {
		result1 *protos.Credentials
		result2 error
	}
	queryEnclaveCredentialsReturnsOnCall map[int]struct {
		result1 *protos.Credentials
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ErccStub) QueryEnclaveCredentials(arg1 shim.ChaincodeStubInterface, arg2 string, arg3 string, arg4 string) (*protos.Credentials, error) {
	fake.queryEnclaveCredentialsMutex.Lock()
	ret, specificReturn := fake.queryEnclaveCredentialsReturnsOnCall[len(fake.queryEnclaveCredentialsArgsForCall)]
	fake.queryEnclaveCredentialsArgsForCall = append(fake.queryEnclaveCredentialsArgsForCall, struct {
		arg1 shim.ChaincodeStubInterface
		arg2 string
		arg3 string
		arg4 string
	}{arg1, arg2, arg3, arg4})
	stub := fake.QueryEnclaveCredentialsStub
	fakeReturns := fake.queryEnclaveCredentialsReturns
	fake.recordInvocation("QueryEnclaveCredentials", []interface{}{arg1, arg2, arg3, arg4})
	fake.queryEnclaveCredentialsMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ErccStub) QueryEnclaveCredentialsCallCount() int {
	fake.queryEnclaveCredentialsMutex.RLock()
	defer fake.queryEnclaveCredentialsMutex.RUnlock()
	return len(fake.queryEnclaveCredentialsArgsForCall)
}

func (fake *ErccStub) QueryEnclaveCredentialsCalls(stub func(shim.ChaincodeStubInterface, string, string, string) (*protos.Credentials, error)) {
	fake.queryEnclaveCredentialsMutex.Lock()
	defer fake.queryEnclaveCredentialsMutex.Unlock()
	fake.QueryEnclaveCredentialsStub = stub
}

func (fake *ErccStub) QueryEnclaveCredentialsArgsForCall(i int) (shim.ChaincodeStubInterface, string, string, string) {
	fake.queryEnclaveCredentialsMutex.RLock()
	defer fake.queryEnclaveCredentialsMutex.RUnlock()
	argsForCall := fake.queryEnclaveCredentialsArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *ErccStub) QueryEnclaveCredentialsReturns(result1 *protos.Credentials, result2 error) {
	fake.queryEnclaveCredentialsMutex.Lock()
	defer fake.queryEnclaveCredentialsMutex.Unlock()
	fake.QueryEnclaveCredentialsStub = nil
	fake.queryEnclaveCredentialsReturns = struct {
		result1 *protos.Credentials
		result2 error
	}{result1, result2}
}

func (fake *ErccStub) QueryEnclaveCredentialsReturnsOnCall(i int, result1 *protos.Credentials, result2 error) {
	fake.queryEnclaveCredentialsMutex.Lock()
	defer fake.queryEnclaveCredentialsMutex.Unlock()
	fake.QueryEnclaveCredentialsStub = nil
	if fake.queryEnclaveCredentialsReturnsOnCall == nil {
		fake.queryEnclaveCredentialsReturnsOnCall = make(map[int]struct {
			result1 *protos.Credentials
			result2 error
		})
	}
	fake.queryEnclaveCredentialsReturnsOnCall[i] = struct {
		result1 *protos.Credentials
		result2 error
	}{result1, result2}
}

func (fake *ErccStub) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.queryEnclaveCredentialsMutex.RLock()
	defer fake.queryEnclaveCredentialsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ErccStub) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
