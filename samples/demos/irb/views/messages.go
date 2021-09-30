package views

import "encoding/json"

type ApprovalRequestNotification struct {
	Message      string
	Sender       string
	ExperimentID string
}

func (a ApprovalRequestNotification) Serialize() ([]byte, error) {
	return json.Marshal(a)
}
