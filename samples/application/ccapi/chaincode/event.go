package chaincode

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"

	"github.com/hyperledger-labs/ccapi/common"
	ev "github.com/hyperledger/fabric-sdk-go/pkg/client/event"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
)

func getEventClient(channelName string) (*ev.Client, error) {
	// create channel manager
	fabMngr, err := common.NewFabricChClient(channelName, os.Getenv("USER"), os.Getenv("ORG"))
	if err != nil {
		return nil, err
	}

	// Create event client
	ec, err := ev.New(fabMngr.Provider, ev.WithBlockEvents())
	if err != nil {
		return nil, err
	}

	return ec, nil
}

func WaitForEvent(channelName, ccName, eventName string, fn func(*fab.CCEvent)) {
	ec, err := getEventClient(channelName)
	if err != nil {
		log.Println("error getting event client: ", err)
		return
	}

	for {
		// Register chaincode event
		registration, notifier, err := ec.RegisterChaincodeEvent(ccName, eventName)
		if err != nil {
			log.Println("error registering chaincode event: ", err)
			return
		}

		// Execute handler function on event notification
		ccEvent := <-notifier
		fmt.Printf("Received CC event: %v\n", ccEvent)
		fn(ccEvent)

		ec.Unregister(registration)
	}
}

func HandleEvent(channelName, ccName string, event EventHandler) {
	ec, err := getEventClient(channelName)
	if err != nil {
		log.Println("error getting event client: ", err)
		return
	}

	for {
		// Register chaincode event
		registration, notifier, err := ec.RegisterChaincodeEvent(ccName, event.Tag)
		if err != nil {
			log.Println("error registering chaincode event: ", err)
			return
		}

		// Execute handler function on event notification
		ccEvent := <-notifier
		fmt.Printf("Received CC event: %v\n", ccEvent)
		event.Execute(ccEvent)

		ec.Unregister(registration)
	}
}

func RegisterForEvents() {
	// Get registered events on the chaincode
	res, _, err := Invoke(os.Getenv("CHANNEL"), os.Getenv("CCNAME"), "getEvents", os.Getenv("USER"), nil, nil)
	if err != nil {
		fmt.Println("error registering for events: ", err)
		return
	}

	var events []interface{}
	nerr := json.Unmarshal(res.Payload, &events)
	if nerr != nil {
		fmt.Println("error unmarshalling events: ", nerr)
		return
	}

	msp := common.GetClientOrg() + "MSP"

	for _, event := range events {
		eventMap := event.(map[string]interface{})
		receiverArr, ok := eventMap["receivers"]

		isReceiver := true
		// Verify if the MSP is a receiver for the event
		if ok {
			isReceiver = false
			receivers := receiverArr.([]interface{})
			for _, r := range receivers {
				receiver := r.(string)

				if len(receiver) <= 1 {
					continue
				}
				if receiver[0] == '$' {
					match, err := regexp.MatchString(receiver[1:], msp)
					if err != nil {
						fmt.Println("error matching regexp: ", err)
						return
					}
					if match {
						isReceiver = true
						break
					}
				} else {
					if receiver == msp {
						isReceiver = true
						break
					}
				}
			}
		}

		if isReceiver {
			eventHandler := EventHandler{
				Tag:         eventMap["tag"].(string),
				Type:        EventType(eventMap["type"].(float64)),
				Transaction: eventMap["transaction"].(string),
				Channel:     eventMap["channel"].(string),
				Chaincode:   eventMap["chaincode"].(string),
				Label:       eventMap["label"].(string),
				BaseLog:     eventMap["baseLog"].(string),
				ReadOnly:    eventMap["readOnly"].(bool),
			}

			go HandleEvent(os.Getenv("CHANNEL"), os.Getenv("CCNAME"), eventHandler)
		}
	}
}
