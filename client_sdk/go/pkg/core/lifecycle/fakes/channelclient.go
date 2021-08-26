// Code generated by counterfeiter. DO NOT EDIT.
package fakes

import (
	"sync"
)

type ChannelClient struct {
	ExecuteStub        func(string, string, [][]byte) (string, error)
	executeMutex       sync.RWMutex
	executeArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 [][]byte
	}
	executeReturns struct {
		result1 string
		result2 error
	}
	executeReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	QueryStub        func(string, string, [][]byte, ...string) ([]byte, error)
	queryMutex       sync.RWMutex
	queryArgsForCall []struct {
		arg1 string
		arg2 string
		arg3 [][]byte
		arg4 []string
	}
	queryReturns struct {
		result1 []byte
		result2 error
	}
	queryReturnsOnCall map[int]struct {
		result1 []byte
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ChannelClient) Execute(arg1 string, arg2 string, arg3 [][]byte) (string, error) {
	var arg3Copy [][]byte
	if arg3 != nil {
		arg3Copy = make([][]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.executeMutex.Lock()
	ret, specificReturn := fake.executeReturnsOnCall[len(fake.executeArgsForCall)]
	fake.executeArgsForCall = append(fake.executeArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 [][]byte
	}{arg1, arg2, arg3Copy})
	stub := fake.ExecuteStub
	fakeReturns := fake.executeReturns
	fake.recordInvocation("Execute", []interface{}{arg1, arg2, arg3Copy})
	fake.executeMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ChannelClient) ExecuteCallCount() int {
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
	return len(fake.executeArgsForCall)
}

func (fake *ChannelClient) ExecuteCalls(stub func(string, string, [][]byte) (string, error)) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = stub
}

func (fake *ChannelClient) ExecuteArgsForCall(i int) (string, string, [][]byte) {
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
	argsForCall := fake.executeArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *ChannelClient) ExecuteReturns(result1 string, result2 error) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = nil
	fake.executeReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ChannelClient) ExecuteReturnsOnCall(i int, result1 string, result2 error) {
	fake.executeMutex.Lock()
	defer fake.executeMutex.Unlock()
	fake.ExecuteStub = nil
	if fake.executeReturnsOnCall == nil {
		fake.executeReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.executeReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *ChannelClient) Query(arg1 string, arg2 string, arg3 [][]byte, arg4 ...string) ([]byte, error) {
	var arg3Copy [][]byte
	if arg3 != nil {
		arg3Copy = make([][]byte, len(arg3))
		copy(arg3Copy, arg3)
	}
	fake.queryMutex.Lock()
	ret, specificReturn := fake.queryReturnsOnCall[len(fake.queryArgsForCall)]
	fake.queryArgsForCall = append(fake.queryArgsForCall, struct {
		arg1 string
		arg2 string
		arg3 [][]byte
		arg4 []string
	}{arg1, arg2, arg3Copy, arg4})
	stub := fake.QueryStub
	fakeReturns := fake.queryReturns
	fake.recordInvocation("Query", []interface{}{arg1, arg2, arg3Copy, arg4})
	fake.queryMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3, arg4...)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ChannelClient) QueryCallCount() int {
	fake.queryMutex.RLock()
	defer fake.queryMutex.RUnlock()
	return len(fake.queryArgsForCall)
}

func (fake *ChannelClient) QueryCalls(stub func(string, string, [][]byte, ...string) ([]byte, error)) {
	fake.queryMutex.Lock()
	defer fake.queryMutex.Unlock()
	fake.QueryStub = stub
}

func (fake *ChannelClient) QueryArgsForCall(i int) (string, string, [][]byte, []string) {
	fake.queryMutex.RLock()
	defer fake.queryMutex.RUnlock()
	argsForCall := fake.queryArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3, argsForCall.arg4
}

func (fake *ChannelClient) QueryReturns(result1 []byte, result2 error) {
	fake.queryMutex.Lock()
	defer fake.queryMutex.Unlock()
	fake.QueryStub = nil
	fake.queryReturns = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *ChannelClient) QueryReturnsOnCall(i int, result1 []byte, result2 error) {
	fake.queryMutex.Lock()
	defer fake.queryMutex.Unlock()
	fake.QueryStub = nil
	if fake.queryReturnsOnCall == nil {
		fake.queryReturnsOnCall = make(map[int]struct {
			result1 []byte
			result2 error
		})
	}
	fake.queryReturnsOnCall[i] = struct {
		result1 []byte
		result2 error
	}{result1, result2}
}

func (fake *ChannelClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.executeMutex.RLock()
	defer fake.executeMutex.RUnlock()
	fake.queryMutex.RLock()
	defer fake.queryMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ChannelClient) recordInvocation(key string, args []interface{}) {
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
