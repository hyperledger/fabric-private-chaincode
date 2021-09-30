package experimenter

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
)

type Execution struct {
	ExperimentId string
}

type ExecutionView struct {
	*Execution
}

func (c *ExecutionView) Call(context view.Context) (interface{}, error) {
	fmt.Println("All cool, now we can run our experimenter")

	resp, err := http.Get("http://localhost:5000/attestation")
	if err != nil {
		return nil, err
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	fmt.Printf("got: %s\n", bodyBytes)

	return nil, nil
}

func NewExecutionView(ExperimentID string) view.View {
	return &ExecutionView{
		Execution: &Execution{ExperimentId: ExperimentID},
	}
}
