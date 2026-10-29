/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package gateway

import (
	"fmt"
	"strings"

	gatewaypb "github.com/hyperledger/fabric-protos-go-apiv2/gateway"
	"google.golang.org/grpc/status"
)

// WithErrorDetails appends the per-peer error details attached by the Fabric Gateway to the error message.
// The Fabric Gateway only returns a generic message (e.g., "failed to evaluate transaction") and reports the
// actual chaincode errors as gRPC status details.
func WithErrorDetails(err error) error {
	if err == nil {
		return nil
	}

	var details []string
	for _, detail := range status.Convert(err).Details() {
		if d, ok := detail.(*gatewaypb.ErrorDetail); ok {
			details = append(details, fmt.Sprintf("%s (%s): %s", d.GetAddress(), d.GetMspId(), d.GetMessage()))
		}
	}
	if len(details) == 0 {
		return err
	}

	return fmt.Errorf("%w [details: %s]", err, strings.Join(details, "; "))
}
