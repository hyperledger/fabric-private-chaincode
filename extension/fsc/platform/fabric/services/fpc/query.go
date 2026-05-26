/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fpc

import (
	"strconv"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/pkg/errors"
)

// ChaincodeQuery models the query of an FPC
type ChaincodeQuery struct {
	*Chaincode

	function string
	args     []interface{}
}

func (i *ChaincodeQuery) WithSignerIdentity(identity view.Identity) *ChaincodeQuery {
	return i
}

func (i *ChaincodeQuery) WithTransientEntry(k string, v interface{}) {
	panic("implement me")
}

func (i *ChaincodeQuery) WithEndorsers(endorsers ...view.Identity) {
	panic("implement me")
}

func (i *ChaincodeQuery) WithEndorsersByMSPIDs(ds ...string) {
	panic("implement me")
}

func (i *ChaincodeQuery) WithEndorsersFromMyOrg() {
	panic("implement me")
}

// Call invokes the chaincode and returns the result
func (i *ChaincodeQuery) Call() ([]byte, error) {
	logger.Debugf("query FPC [%s:%s:%s]", i.Channel, i.ID, i.function)
	args, err := i.prepareArgs()
	if err != nil {
		return nil, errors.WithMessagef(err, "failed preparing arguments")
	}
	return i.InvokerContract.EvaluateTransaction(i.function, args...)
}

func (i *ChaincodeQuery) prepareArgs() ([]string, error) {
	var args []string
	for _, arg := range i.args {
		b, err := i.toString(arg)
		if err != nil {
			return nil, err
		}
		args = append(args, b)
	}
	return args, nil
}

func (i *ChaincodeQuery) toString(arg interface{}) (string, error) {
	switch v := arg.(type) {
	case []byte:
		return string(v), nil
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	default:
		return "", errors.Errorf("arg type [%T] not recognized.", v)
	}
}

func (i *ChaincodeQuery) WithMatchEndorsementPolicy() {
}
