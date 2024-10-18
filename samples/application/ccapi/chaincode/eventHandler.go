package chaincode

import (
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

type EventType float64

const (
	EventLog EventType = iota
	EventTransaction
	EventCustom
)

type EventHandler struct {
	Tag         string
	Label       string
	Type        EventType
	Transaction string
	Channel     string
	Chaincode   string
	BaseLog     string
	ReadOnly    bool
}

func (event EventHandler) Execute(ccEvent *fab.CCEvent) {
	if len(event.BaseLog) > 0 {
		fmt.Println(event.BaseLog)
	}

	if event.Type == EventLog {
		var logStr string
		nerr := json.Unmarshal(ccEvent.Payload, &logStr)
		if nerr != nil {
			fmt.Println("error unmarshalling log: ", nerr)
			return
		}

		if len(logStr) > 0 {
			fmt.Println("Event '", event.Label, "' log: ", logStr)
		}
	} else if event.Type == EventTransaction {
		ch := os.Getenv("CHANNEL")
		if event.Channel != "" {
			ch = event.Channel
		}
		cc := os.Getenv("CCNAME")
		if event.Chaincode != "" {
			cc = event.Chaincode
		}

		res, _, err := Invoke(ch, cc, event.Transaction, os.Getenv("USER"), [][]byte{ccEvent.Payload}, nil)
		if err != nil {
			fmt.Println("error invoking transaction: ", err)
			return
		}

		var response map[string]interface{}
		nerr := json.Unmarshal(res.Payload, &response)
		if nerr != nil {
			fmt.Println("error unmarshalling response: ", nerr)
			return
		}
		fmt.Println("Response: ", response)
	} else if event.Type == EventCustom {
		// Encode payload to base64
		b64Encode := b64.StdEncoding.EncodeToString([]byte(ccEvent.Payload))

		args, ok := json.Marshal(map[string]interface{}{
			"eventTag": event.Tag,
			"payload":  b64Encode,
		})
		if ok != nil {
			fmt.Println("failed to encode args to JSON format")
			return
		}

		// Invoke tx
		txName := "executeEvent"
		if event.ReadOnly {
			txName = "runEvent"
		}

		_, _, err := Invoke(os.Getenv("CHANNEL"), os.Getenv("CCNAME"), txName, os.Getenv("USER"), [][]byte{args}, nil)
		if err != nil {
			fmt.Println("error invoking transaction: ", err)
			return
		}
	} else {
		fmt.Println("Event type not supported")
	}
}
