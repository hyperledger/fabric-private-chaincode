// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"

	"github.com/hyperledger-labs/fabric-private-chaincode/client_sdk/go/pkg/gateway/internal"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	gatewaya "github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)

type Contract struct {
	CreateTransactionStub        func(string, ...func(*gatewaya.Transaction) error) (internal.Transaction, error)
	createTransactionMutex       sync.RWMutex
	createTransactionArgsForCall []struct {
		arg1 string
		arg2 []func(*gatewaya.Transaction) error
	}
	createTransactionReturns struct {
		result1 internal.Transaction
		result2 error
	}
	createTransactionReturnsOnCall map[int]struct {
		result1 internal.Transaction
		result2 error
	}
	EvaluateTransactionStub        func(string, ...string) ([]byte, error)
	evaluateTransactionMutex       sync.RWMutex
	evaluateTransactionArgsForCall []struct {
		arg1 string
		arg2 []string
	}
	evaluateTransactionReturns struct {
		result1 []byte
		result2 error
	}
	evaluateTransactionReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	NameStub        func() string
	nameMutex       sync.RWMutex
	nameArgsForCall []struct {
	}
	nameReturns struct {
		result1 string
	}
	nameReturnsOnCall map[int]struct {
		result1 string
	}
	RegisterEventStub        func(string) (fab.Registration, <-chan *fab.CCEvent, error)
	registerEventMutex       sync.RWMutex
	registerEventArgsForCall []struct {
		arg1 string
	}
	registerEventReturns struct {
		result1 fab.Registration
		result2 <-chan *fab.CCEvent
		result3 error
	}
	registerEventReturnsOnCall map[int]struct {
		result1 fab.Registration
		result2 <-chan *fab.CCEvent
		result3 error
	}
	SubmitTransactionStub        func(string, ...string) ([]byte, error)
	submitTransactionMutex       sync.RWMutex
	submitTransactionArgsForCall []struct {
		arg1 string
		arg2 []string
	}
	submitTransactionReturns struct {
		result1 []byte
		result2 error
	}
	submitTransactionReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	UnregisterStub        func(fab.Registration)
	unregisterMutex       sync.RWMutex
	unregisterArgsForCall []struct {
		arg1 fab.Registration
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Contract) CreateTransaction(arg1 string, arg2 ...func(*gatewaya.Transaction) error) (internal.Transaction, error) {
	fake.createTransactionMutex.Lock()
	ret, specificReturn := fake.createTransactionReturnsOnCall[len(fake.createTransactionArgsForCall)]
	fake.createTransactionArgsForCall = append(fake.createTransactionArgsForCall, struct {
		arg1 string
		arg2 []func(*gatewaya.Transaction) error
	}{arg1, arg2})
	stub := fake.CreateTransactionStub
	fakeReturns := fake.createTransactionReturns
	fake.recordInvocation("CreateTransaction", []interface{}{arg1, arg2})
	fake.createTransactionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Contract) CreateTransactionCallCount() int {
	fake.createTransactionMutex.RLock()
	defer fake.createTransactionMutex.RUnlock()
	return len(fake.createTransactionArgsForCall)
}

func (fake *Contract) CreateTransactionCalls(stub func(string, ...func(*gatewaya.Transaction) error) (internal.Transaction, error)) {
	fake.createTransactionMutex.Lock()
	defer fake.createTransactionMutex.Unlock()
	fake.CreateTransactionStub = stub
}

func (fake *Contract) CreateTransactionArgsForCall(i int) (string, []func(*gatewaya.Transaction) error) {
	fake.createTransactionMutex.RLock()
	defer fake.createTransactionMutex.RUnlock()
	argsForCall := fake.createTransactionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Contract) CreateTransactionReturns(result1 internal.Transaction, result2 error) {
	fake.createTransactionMutex.Lock()
	defer fake.createTransactionMutex.Unlock()
	fake.CreateTransactionStub = nil
	fake.createTransactionReturns = struct {
		result1 internal.Transaction
		result2 error
	}{result1, result2}
}

func (fake *Contract) CreateTransactionReturnsOnCall(i int, result1 internal.Transaction, result2 error) {
	fake.createTransactionMutex.Lock()
	defer fake.createTransactionMutex.Unlock()
	fake.CreateTransactionStub = nil
	if fake.createTransactionReturnsOnCall == nil {
		fake.createTransactionReturnsOnCall = make(map[int]struct {
			result1 internal.Transaction
			result2 error
		})
	}
	fake.createTransactionReturnsOnCall[i] = struct {
		result1 internal.Transaction
		result2 error
	}{result1, result2}
}

func (fake *Contract) EvaluateTransaction(arg1 string, arg2 ...string) ([]byte, error) {
	fake.evaluateTransactionMutex.Lock()
	ret, specificReturn := fake.evaluateTransactionReturnsOnCall[len(fake.evaluateTransactionArgsForCall)]
	fake.evaluateTransactionArgsForCall = append(fake.evaluateTransactionArgsForCall, struct {
		arg1 string
		arg2 []string
	}{arg1, arg2})
	stub := fake.EvaluateTransactionStub
	fakeReturns := fake.evaluateTransactionReturns
	fake.recordInvocation("EvaluateTransaction", []interface{}{arg1, arg2})
	fake.evaluateTransactionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Contract) EvaluateTransactionCallCount() int {
	fake.evaluateTransactionMutex.RLock()
	defer fake.evaluateTransactionMutex.RUnlock()
	return len(fake.evaluateTransactionArgsForCall)
}

func (fake *Contract) EvaluateTransactionCalls(stub func(string, ...string) ([]byte, error)) {
	fake.evaluateTransactionMutex.Lock()
	defer fake.evaluateTransactionMutex.Unlock()
	fake.EvaluateTransactionStub = stub
}

func (fake *Contract) EvaluateTransactionArgsForCall(i int) (string, []string) {
	fake.evaluateTransactionMutex.RLock()
	defer fake.evaluateTransactionMutex.RUnlock()
	argsForCall := fake.evaluateTransactionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Contract) EvaluateTransactionReturns(result1 []byte, result2 error) {
	fake.evaluateTransactionMutex.Lock()
	defer fake.evaluateTransactionMutex.Unlock()
	fake.EvaluateTransactionStub = nil
	fake.evaluateTransactionReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *Contract) EvaluateTransactionReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.evaluateTransactionMutex.Lock()
	defer fake.evaluateTransactionMutex.Unlock()
	fake.EvaluateTransactionStub = nil
	if fake.evaluateTransactionReturnsOnCall == nil {
		fake.evaluateTransactionReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.evaluateTransactionReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *Contract) Name() string {
	fake.nameMutex.Lock()
	ret, specificReturn := fake.nameReturnsOnCall[len(fake.nameArgsForCall)]
	fake.nameArgsForCall = append(fake.nameArgsForCall, struct {
	}{})
	stub := fake.NameStub
	fakeReturns := fake.nameReturns
	fake.recordInvocation("Name", []interface{}{})
	fake.nameMutex.Unlock()
	if stub != nil {
		return stub()
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *Contract) NameCallCount() int {
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	return len(fake.nameArgsForCall)
}

func (fake *Contract) NameCalls(stub func() string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = stub
}

func (fake *Contract) NameReturns(result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	fake.nameReturns = struct {
		result1 string
	}{result1}
}

func (fake *Contract) NameReturnsOnCall(i int, result1 string) {
	fake.nameMutex.Lock()
	defer fake.nameMutex.Unlock()
	fake.NameStub = nil
	if fake.nameReturnsOnCall == nil {
		fake.nameReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.nameReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *Contract) RegisterEvent(arg1 string) (fab.Registration, <-chan *fab.CCEvent, error) {
	fake.registerEventMutex.Lock()
	ret, specificReturn := fake.registerEventReturnsOnCall[len(fake.registerEventArgsForCall)]
	fake.registerEventArgsForCall = append(fake.registerEventArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.RegisterEventStub
	fakeReturns := fake.registerEventReturns
	fake.recordInvocation("RegisterEvent", []interface{}{arg1})
	fake.registerEventMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2, ret.result3
	}
	return fakeReturns.result1, fakeReturns.result2, fakeReturns.result3
}

func (fake *Contract) RegisterEventCallCount() int {
	fake.registerEventMutex.RLock()
	defer fake.registerEventMutex.RUnlock()
	return len(fake.registerEventArgsForCall)
}

func (fake *Contract) RegisterEventCalls(stub func(string) (fab.Registration, <-chan *fab.CCEvent, error)) {
	fake.registerEventMutex.Lock()
	defer fake.registerEventMutex.Unlock()
	fake.RegisterEventStub = stub
}

func (fake *Contract) RegisterEventArgsForCall(i int) string {
	fake.registerEventMutex.RLock()
	defer fake.registerEventMutex.RUnlock()
	argsForCall := fake.registerEventArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Contract) RegisterEventReturns(result1 fab.Registration, result2 <-chan *fab.CCEvent, result3 error) {
	fake.registerEventMutex.Lock()
	defer fake.registerEventMutex.Unlock()
	fake.RegisterEventStub = nil
	fake.registerEventReturns = struct {
		result1 fab.Registration
		result2 <-chan *fab.CCEvent
		result3 error
	}{result1, result2, result3}
}

func (fake *Contract) RegisterEventReturnsOnCall(i int, result1 fab.Registration, result2 <-chan *fab.CCEvent, result3 error) {
	fake.registerEventMutex.Lock()
	defer fake.registerEventMutex.Unlock()
	fake.RegisterEventStub = nil
	if fake.registerEventReturnsOnCall == nil {
		fake.registerEventReturnsOnCall = make(map[int]struct {
			result1 fab.Registration
			result2 <-chan *fab.CCEvent
			result3 error
		})
	}
	fake.registerEventReturnsOnCall[i] = struct {
		result1 fab.Registration
		result2 <-chan *fab.CCEvent
		result3 error
	}{result1, result2, result3}
}

func (fake *Contract) SubmitTransaction(arg1 string, arg2 ...string) ([]byte, error) {
	fake.submitTransactionMutex.Lock()
	ret, specificReturn := fake.submitTransactionReturnsOnCall[len(fake.submitTransactionArgsForCall)]
	fake.submitTransactionArgsForCall = append(fake.submitTransactionArgsForCall, struct {
		arg1 string
		arg2 []string
	}{arg1, arg2})
	stub := fake.SubmitTransactionStub
	fakeReturns := fake.submitTransactionReturns
	fake.recordInvocation("SubmitTransaction", []interface{}{arg1, arg2})
	fake.submitTransactionMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Contract) SubmitTransactionCallCount() int {
	fake.submitTransactionMutex.RLock()
	defer fake.submitTransactionMutex.RUnlock()
	return len(fake.submitTransactionArgsForCall)
}

func (fake *Contract) SubmitTransactionCalls(stub func(string, ...string) ([]byte, error)) {
	fake.submitTransactionMutex.Lock()
	defer fake.submitTransactionMutex.Unlock()
	fake.SubmitTransactionStub = stub
}

func (fake *Contract) SubmitTransactionArgsForCall(i int) (string, []string) {
	fake.submitTransactionMutex.RLock()
	defer fake.submitTransactionMutex.RUnlock()
	argsForCall := fake.submitTransactionArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *Contract) SubmitTransactionReturns(result1 []byte, result2 error) {
	fake.submitTransactionMutex.Lock()
	defer fake.submitTransactionMutex.Unlock()
	fake.SubmitTransactionStub = nil
	fake.submitTransactionReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *Contract) SubmitTransactionReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.submitTransactionMutex.Lock()
	defer fake.submitTransactionMutex.Unlock()
	fake.SubmitTransactionStub = nil
	if fake.submitTransactionReturnsOnCall == nil {
		fake.submitTransactionReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.submitTransactionReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *Contract) Unregister(arg1 fab.Registration) {
	fake.unregisterMutex.Lock()
	fake.unregisterArgsForCall = append(fake.unregisterArgsForCall, struct {
		arg1 fab.Registration
	}{arg1})
	stub := fake.UnregisterStub
	fake.recordInvocation("Unregister", []interface{}{arg1})
	fake.unregisterMutex.Unlock()
	if stub != nil {
		fake.UnregisterStub(arg1)
	}
}

func (fake *Contract) UnregisterCallCount() int {
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
	return len(fake.unregisterArgsForCall)
}

func (fake *Contract) UnregisterCalls(stub func(fab.Registration)) {
	fake.unregisterMutex.Lock()
	defer fake.unregisterMutex.Unlock()
	fake.UnregisterStub = stub
}

func (fake *Contract) UnregisterArgsForCall(i int) fab.Registration {
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
	argsForCall := fake.unregisterArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Contract) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.createTransactionMutex.RLock()
	defer fake.createTransactionMutex.RUnlock()
	fake.evaluateTransactionMutex.RLock()
	defer fake.evaluateTransactionMutex.RUnlock()
	fake.nameMutex.RLock()
	defer fake.nameMutex.RUnlock()
	fake.registerEventMutex.RLock()
	defer fake.registerEventMutex.RUnlock()
	fake.submitTransactionMutex.RLock()
	defer fake.submitTransactionMutex.RUnlock()
	fake.unregisterMutex.RLock()
	defer fake.unregisterMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Contract) recordInvocation(key string, args []interface{}) {
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