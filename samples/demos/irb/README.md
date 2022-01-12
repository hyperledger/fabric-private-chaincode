# Institutional Review Board (IRB) Sample

This demo implements an FPC-based IRB experiment approval services to protect the confidentiality of data being used in analytical experiments. 
The problem addressed here is relevant in a healthcare context, where sensitive patient data is processed in clinical trails.    
Multiple participants, namely, data providers, experimenter, and principal investigators are collaborating to conduct analytical research using sensitive data.  

We show a prototype of an application that performs analytical experiments on data such that:
- Any constraints on the use of the data are respected (consent).
- The confidentiality of data and policies for its use is maintained.
- Analytical experiments must be approved by an IRB which ensures that the description and implementation of the experiment are appropriate.
- Analytical experiments are executed in an SGX container (Graphene) that can attest to the integrity of the computation.

The demo was presented at the Hyperledger Global Forum (HLGF) 2021. Check the recordings on [youtube](https://www.youtube.com/watch?v=MU4BpZp8A1Y).

Note that the demo source code here is a simplified version of the demo presented at HLGF21, to fully focus on the FPC implementation.
In particular, the experiment is not protected with SGX and the WebUI component is not present.
The full source code for the HLFG21 presentation is located in the [irb-demo](https://github.com/hyperledger/fabric-private-chaincode/tree/irb-demo/samples/demos/irb) branch.

The IRB use case requires the interaction between the participants. 
A typical application flow begins with the data registration and consent, following the approval protocol, and finally executing the experiment and publishing the results.
We use [Fabric Smart Client](https://github.com/hyperledger-labs/fabric-smart-client) to implement the complex interactions. 

## What the demo shows

The demo runs in the terminal and demonstrates the application flow.
In particular, the demo starts a Fabric network with multiple organizations participating and hosting fabric peers.
The Fabric Smart Client installs the FPC Chaincode that implements the IRB experiment approval service.

Once the network is ready, the application flow begins with the investigator, creating a new study.
Next, the data provider registers and uploads new patient data to the systems, assigning the data to be used in the study.
Now we have a study and data to process.

The experimenter proceeds with creating a new experiment and asks the investigator for approval.
Once the approval arrived, that is, the investigator has reviewed the experiment proposal and submitted an approval to the IRB experiment approval service,  the execution of the experiment is triggered.

Finally, the experimenter requests an evaluation pack from the IRB experiment approval service, that contains the access information to collect the patient data.
The evaluation pack is passed to the experiment instance, which fetches the patient data, performs the experiment computation, and returns the result.
Note that the patient data is encrypted in a way that only an approved experiment instance can decrypt the data.
More details on that, see the HLGF21 presentation.

## Code structure
- `/chaincode` contains the FPC chaincode that implements the IRV approval service
- `/experimenter` contains the analytic experiment code based on [PyTorch](https://pytorch.org/).
- `/pkg` contains implementations for various components, including crypto, container management, etc. 
- `/protos` contains the message definitions used in this application
- `/views` contains the protocol implementation for the data provider, experimenter, and investigator.
- `irb_test.go` is the starting point of the demo.
- `topology.go` defines the Fabric network used in this demo.

## Setup

The demo uses redis to store the data provided by the patients. Get a redis docker image.
```bash
docker pull redis:latest
```

Next, we build the components of the demo by running:
```bash
make build
```

## Run the demo

To run the demo just use the `test` target.
You will see the output of the interaction between the participants in your terminal.

```bash
make test
```
