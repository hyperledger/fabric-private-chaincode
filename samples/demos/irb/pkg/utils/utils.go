/*
   Copyright IBM Corp. All Rights Reserved.

   SPDX-License-Identifier: Apache-2.0
*/

package utils

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	pb "github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/protos"
	"google.golang.org/protobuf/proto"
)

func MarshalProtoBase64(msg proto.Message) string {
	bytes, _ := proto.Marshal(msg)
	return base64.StdEncoding.EncodeToString(bytes)
}

func Retry(f func() bool, maxAttempt int, maxTimeout time.Duration, delay time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), maxTimeout)
	defer cancel()
	for i := 0; i < maxAttempt; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if f() {
				return nil
			}
			time.Sleep(delay)
			delay *= 2
		}
	}

	return fmt.Errorf("max attempts reached")
}

func UnmarshalStatus(statusBytes []byte) (*pb.Status, error) {
	status := &pb.Status{}
	err := proto.Unmarshal(statusBytes, status)
	if err != nil {
		return nil, err
	}

	return status, nil
}
