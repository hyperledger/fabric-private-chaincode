/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package messages

import "encoding/json"

type ApprovalRequestNotification struct {
	Message      string
	Sender       string
	ExperimentID string
}

func (a ApprovalRequestNotification) Serialize() ([]byte, error) {
	return json.Marshal(a)
}
