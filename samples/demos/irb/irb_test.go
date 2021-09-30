package irb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/hyperledger-labs/fabric-smart-client/integration"
	"github.com/hyperledger-labs/fabric-smart-client/integration/nwo/common"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/container"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/pdf"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/pkg/users"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/dataprovider"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/experimenter"
	"github.com/hyperledger/fabric-private-chaincode/samples/demos/irb/views/investigator"
	"github.com/stretchr/testify/assert"
)

const studyID = "pineapple"
const experimentID = "exp001"

func TestFlow(t *testing.T) {

	// create test patients
	patients := []string{"patient1", "patient2"}
	fmt.Printf("Creating patients and data...")
	for i := 0; i < len(patients); i++ {
		_, _, _, err := users.LoadOrCreateUser(patients[i])
		assert.NoError(t, err)
	}
	fmt.Printf("done.\n")

	// setup network
	ii, err := integration.Generate(23000, Topology()...)
	assert.NoError(t, err)
	ii.Start()
	defer ii.Stop()

	// register new study
	_, err = ii.Client("investigator").CallView("RegisterStudy", common.JSONMarshall(&investigator.RegisterStudy{
		StudyID:  studyID,
		Metadata: "some fancy study",
	}))
	assert.NoError(t, err)

	// setup redis
	redis := &container.Container{
		Image:    "redis",
		Name:     "redis-container",
		HostIP:   "localhost",
		HostPort: "6379",
	}
	err = redis.Start()
	assert.NoError(t, err)
	defer redis.Stop()

	// setup experiment container
	experiment := &container.Container{
		Image:    "irb-experimenter-worker",
		Name:     "experiment-container",
		HostIP:   "localhost",
		HostPort: "5000",
	}
	err = experiment.Start()
	defer experiment.Stop()
	assert.NoError(t, err)
	time.Sleep(10 * time.Second)

	resp, err := http.Get("http://localhost:5000/info")
	assert.NoError(t, err)

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)

	fmt.Printf("got: %s\n", bodyBytes)

	assert.NotEmpty(t, "stop here")

	// data provider flow
	// parse pdf provided by patient
	data, err := ioutil.ReadFile("pkg/pdf/patient1.pdf")
	res, err := pdf.ParseQuestionForm(data)
	assert.NoError(t, err)

	// load corresponding patientUUID and patientVK
	patientUUID, _, patientVK, err := users.LoadUser(res.UUID)
	assert.NoError(t, err)

	_, err = ii.Client("provider").CallView("RegisterData", common.JSONMarshall(&dataprovider.Register{
		StudyID:     studyID,
		PatientData: []byte(res.Answers.ToString()),
		PatientUUID: string(patientUUID),
		PatientVK:   patientVK,
	}))
	assert.NoError(t, err)

	// experimenter flow
	_, err = ii.Client("experimenter").CallView("Submission", common.JSONMarshall(&experimenter.Submission{
		StudyId:      studyID,
		ExperimentId: experimentID,
		Investigator: ii.Identity("investigator"),
	}))
	assert.NoError(t, err)

}
