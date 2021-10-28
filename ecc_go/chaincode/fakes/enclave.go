// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type EnclaveStub struct {
	ChaincodeInvokeStub        func(shim.ChaincodeStubInterface, []byte) ([]byte, error)
	chaincodeInvokeMutex       sync.RWMutex
	chaincodeInvokeArgsForCall []struct {
		arg1 shim.ChaincodeStubInterface
		arg2 []byte
	}
	chaincodeInvokeReturns struct {
		result1 []byte
		result2 error
	}
	chaincodeInvokeReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	ExportCCKeysStub        func([]byte) ([]byte, error)
	exportCCKeysMutex       sync.RWMutex
	exportCCKeysArgsForCall []struct {
		arg1 []byte
	}
	exportCCKeysReturns struct {
		result1 []byte
		result2 error
	}
	exportCCKeysReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	GenerateCCKeysStub        func() ([]byte, error)
	generateCCKeysMutex       sync.RWMutex
	generateCCKeysArgsForCall []struct {
	}
	generateCCKeysReturns struct {
		result1 []byte
		result2 error
	}
	generateCCKeysReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	GetEnclaveIdStub        func() (string, error)
	getEnclaveIdMutex       sync.RWMutex
	getEnclaveIdArgsForCall []struct {
	}
	getEnclaveIdReturns struct {
		result1 string
		result2 error
	}
	getEnclaveIdReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	ImportCCKeysStub        func() ([]byte, error)
	importCCKeysMutex       sync.RWMutex
	importCCKeysArgsForCall []struct {
	}
	importCCKeysReturns struct {
		result1 []byte
		result2 error
	}
	importCCKeysReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	InitStub        func([]byte, []byte, []byte) ([]byte, error)
	initMutex       sync.RWMutex
	initArgsForCall []struct {
		arg1 []byte
		arg2 []byte
		arg3 []byte
	}
	initReturns struct {
		result1 []byte
		result2 error
	}
	initReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *EnclaveStub) ChaincodeInvoke(arg1 shim.ChaincodeStubInterface, arg2 []byte) ([]byte, error) {
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	fake.chaincodeInvokeMutex.Lock()
	ret, specificReturn := fake.chaincodeInvokeReturnsOnCall[len(fake.chaincodeInvokeArgsForCall)]
	fake.chaincodeInvokeArgsForCall = append(fake.chaincodeInvokeArgsForCall, struct {
		arg1 shim.ChaincodeStubInterface
		arg2 []byte
	}{arg1, arg2Copy})
	stub := fake.ChaincodeInvokeStub
	fakeReturns := fake.chaincodeInvokeReturns
	fake.recordInvocation("ChaincodeInvoke", []interface{}{arg1, arg2Copy})
	fake.chaincodeInvokeMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) ChaincodeInvokeCallCount() int {
	fake.chaincodeInvokeMutex.RLock()
	defer fake.chaincodeInvokeMutex.RUnlock()
	return len(fake.chaincodeInvokeArgsForCall)
}

func (fake *EnclaveStub) ChaincodeInvokeCalls(stub func(shim.ChaincodeStubInterface, []byte) ([]byte, error)) {
	fake.chaincodeInvokeMutex.Lock()
	defer fake.chaincodeInvokeMutex.Unlock()
	fake.ChaincodeInvokeStub = stub
}

func (fake *EnclaveStub) ChaincodeInvokeArgsForCall(i int) (shim.ChaincodeStubInterface, []byte) {
	fake.chaincodeInvokeMutex.RLock()
	defer fake.chaincodeInvokeMutex.RUnlock()
	argsForCall := fake.chaincodeInvokeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *EnclaveStub) ChaincodeInvokeReturns(result1 []byte, result2 error) {
	fake.chaincodeInvokeMutex.Lock()
	defer fake.chaincodeInvokeMutex.Unlock()
	fake.ChaincodeInvokeStub = nil
	fake.chaincodeInvokeReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) ChaincodeInvokeReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.chaincodeInvokeMutex.Lock()
	defer fake.chaincodeInvokeMutex.Unlock()
	fake.ChaincodeInvokeStub = nil
	if fake.chaincodeInvokeReturnsOnCall == nil {
		fake.chaincodeInvokeReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.chaincodeInvokeReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) ExportCCKeys(arg1 []byte) ([]byte, error) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.exportCCKeysMutex.Lock()
	ret, specificReturn := fake.exportCCKeysReturnsOnCall[len(fake.exportCCKeysArgsForCall)]
	fake.exportCCKeysArgsForCall = append(fake.exportCCKeysArgsForCall, struct {
		arg1 []byte
	}{arg1Copy})
	stub := fake.ExportCCKeysStub
	fakeReturns := fake.exportCCKeysReturns
	fake.recordInvocation("ExportCCKeys", []interface{}{arg1Copy})
	fake.exportCCKeysMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) ExportCCKeysCallCount() int {
	fake.exportCCKeysMutex.RLock()
	defer fake.exportCCKeysMutex.RUnlock()
	return len(fake.exportCCKeysArgsForCall)
}

func (fake *EnclaveStub) ExportCCKeysCalls(stub func([]byte) ([]byte, error)) {
	fake.exportCCKeysMutex.Lock()
	defer fake.exportCCKeysMutex.Unlock()
	fake.ExportCCKeysStub = stub
}

func (fake *EnclaveStub) ExportCCKeysArgsForCall(i int) []byte {
	fake.exportCCKeysMutex.RLock()
	defer fake.exportCCKeysMutex.RUnlock()
	argsForCall := fake.exportCCKeysArgsForCall[i]
	return argsForCall.arg1
}

func (fake *EnclaveStub) ExportCCKeysReturns(result1 []byte, result2 error) {
	fake.exportCCKeysMutex.Lock()
	defer fake.exportCCKeysMutex.Unlock()
	fake.ExportCCKeysStub = nil
	fake.exportCCKeysReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) ExportCCKeysReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.exportCCKeysMutex.Lock()
	defer fake.exportCCKeysMutex.Unlock()
	fake.ExportCCKeysStub = nil
	if fake.exportCCKeysReturnsOnCall == nil {
		fake.exportCCKeysReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.exportCCKeysReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) GenerateCCKeys() ([]byte, error) {
	fake.generateCCKeysMutex.Lock()
	ret, specificReturn := fake.generateCCKeysReturnsOnCall[len(fake.generateCCKeysArgsForCall)]
	fake.generateCCKeysArgsForCall = append(fake.generateCCKeysArgsForCall, struct {
	}{})
	stub := fake.GenerateCCKeysStub
	fakeReturns := fake.generateCCKeysReturns
	fake.recordInvocation("GenerateCCKeys", []interface{}{})
	fake.generateCCKeysMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) GenerateCCKeysCallCount() int {
	fake.generateCCKeysMutex.RLock()
	defer fake.generateCCKeysMutex.RUnlock()
	return len(fake.generateCCKeysArgsForCall)
}

func (fake *EnclaveStub) GenerateCCKeysCalls(stub func() ([]byte, error)) {
	fake.generateCCKeysMutex.Lock()
	defer fake.generateCCKeysMutex.Unlock()
	fake.GenerateCCKeysStub = stub
}

func (fake *EnclaveStub) GenerateCCKeysReturns(result1 []byte, result2 error) {
	fake.generateCCKeysMutex.Lock()
	defer fake.generateCCKeysMutex.Unlock()
	fake.GenerateCCKeysStub = nil
	fake.generateCCKeysReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) GenerateCCKeysReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.generateCCKeysMutex.Lock()
	defer fake.generateCCKeysMutex.Unlock()
	fake.GenerateCCKeysStub = nil
	if fake.generateCCKeysReturnsOnCall == nil {
		fake.generateCCKeysReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.generateCCKeysReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) GetEnclaveId() (string, error) {
	fake.getEnclaveIdMutex.Lock()
	ret, specificReturn := fake.getEnclaveIdReturnsOnCall[len(fake.getEnclaveIdArgsForCall)]
	fake.getEnclaveIdArgsForCall = append(fake.getEnclaveIdArgsForCall, struct {
	}{})
	stub := fake.GetEnclaveIdStub
	fakeReturns := fake.getEnclaveIdReturns
	fake.recordInvocation("GetEnclaveId", []interface{}{})
	fake.getEnclaveIdMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) GetEnclaveIdCallCount() int {
	fake.getEnclaveIdMutex.RLock()
	defer fake.getEnclaveIdMutex.RUnlock()
	return len(fake.getEnclaveIdArgsForCall)
}

func (fake *EnclaveStub) GetEnclaveIdCalls(stub func() (string, error)) {
	fake.getEnclaveIdMutex.Lock()
	defer fake.getEnclaveIdMutex.Unlock()
	fake.GetEnclaveIdStub = stub
}

func (fake *EnclaveStub) GetEnclaveIdReturns(result1 string, result2 error) {
	fake.getEnclaveIdMutex.Lock()
	defer fake.getEnclaveIdMutex.Unlock()
	fake.GetEnclaveIdStub = nil
	fake.getEnclaveIdReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) GetEnclaveIdReturnsOnCall(i int, result1 string, result2 error) {
	fake.getEnclaveIdMutex.Lock()
	defer fake.getEnclaveIdMutex.Unlock()
	fake.GetEnclaveIdStub = nil
	if fake.getEnclaveIdReturnsOnCall == nil {
		fake.getEnclaveIdReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.getEnclaveIdReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) ImportCCKeys() ([]byte, error) {
	fake.importCCKeysMutex.Lock()
	ret, specificReturn := fake.importCCKeysReturnsOnCall[len(fake.importCCKeysArgsForCall)]
	fake.importCCKeysArgsForCall = append(fake.importCCKeysArgsForCall, struct {
	}{})
	stub := fake.ImportCCKeysStub
	fakeReturns := fake.importCCKeysReturns
	fake.recordInvocation("ImportCCKeys", []interface{}{})
	fake.importCCKeysMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) ImportCCKeysCallCount() int {
	fake.importCCKeysMutex.RLock()
	defer fake.importCCKeysMutex.RUnlock()
	return len(fake.importCCKeysArgsForCall)
}

func (fake *EnclaveStub) ImportCCKeysCalls(stub func() ([]byte, error)) {
	fake.importCCKeysMutex.Lock()
	defer fake.importCCKeysMutex.Unlock()
	fake.ImportCCKeysStub = stub
}

func (fake *EnclaveStub) ImportCCKeysReturns(result1 []byte, result2 error) {
	fake.importCCKeysMutex.Lock()
	defer fake.importCCKeysMutex.Unlock()
	fake.ImportCCKeysStub = nil
	fake.importCCKeysReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) ImportCCKeysReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.importCCKeysMutex.Lock()
	defer fake.importCCKeysMutex.Unlock()
	fake.ImportCCKeysStub = nil
	if fake.importCCKeysReturnsOnCall == nil {
		fake.importCCKeysReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.importCCKeysReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) Init(arg1 []byte, arg2 []byte, arg3 []byte) ([]byte, error) {
	var arg1Copy []byte
	if arg1 != nil {
		arg1Copy = make([]byte, len(arg1))
		copy(arg1Copy, arg1)
	}
	var arg2Copy []byte
	if arg2 != nil {
		arg2Copy = make([]byte, len(arg2))
		copy(arg2Copy, arg2)
	}
	var arg3Copy []byte
	if arg3 != nil {
		arg3Copy = make([]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.initMutex.Lock()
	ret, specificReturn := fake.initReturnsOnCall[len(fake.initArgsForCall)]
	fake.initArgsForCall = append(fake.initArgsForCall, struct {
		arg1 []byte
		arg2 []byte
		arg3 []byte
	}{arg1Copy, arg2Copy, arg3Copy})
	stub := fake.InitStub
	fakeReturns := fake.initReturns
	fake.recordInvocation("Init", []interface{}{arg1Copy, arg2Copy, arg3Copy})
	fake.initMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *EnclaveStub) InitCallCount() int {
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	return len(fake.initArgsForCall)
}

func (fake *EnclaveStub) InitCalls(stub func([]byte, []byte, []byte) ([]byte, error)) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = stub
}

func (fake *EnclaveStub) InitArgsForCall(i int) ([]byte, []byte, []byte) {
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	argsForCall := fake.initArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *EnclaveStub) InitReturns(result1 []byte, result2 error) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = nil
	fake.initReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) InitReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.initMutex.Lock()
	defer fake.initMutex.Unlock()
	fake.InitStub = nil
	if fake.initReturnsOnCall == nil {
		fake.initReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.initReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *EnclaveStub) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.chaincodeInvokeMutex.RLock()
	defer fake.chaincodeInvokeMutex.RUnlock()
	fake.exportCCKeysMutex.RLock()
	defer fake.exportCCKeysMutex.RUnlock()
	fake.generateCCKeysMutex.RLock()
	defer fake.generateCCKeysMutex.RUnlock()
	fake.getEnclaveIdMutex.RLock()
	defer fake.getEnclaveIdMutex.RUnlock()
	fake.importCCKeysMutex.RLock()
	defer fake.importCCKeysMutex.RUnlock()
	fake.initMutex.RLock()
	defer fake.initMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *EnclaveStub) recordInvocation(key string, args []interface{}) {
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
