# Fabric Private Chaincode

[`Hyperledger Fabric Private Chaincode (FPC)`](https://github.com/hyperledger/fabric-private-chaincode) 
enables the execution of chaincodes using Intel SGX for Hyperledger Fabric.

## Echo 

Echo is a very simple chaincode written in C++ for SGX that returns the name of the function
that has been invoked. In our test, we will make sure that this is what happens.

### Invoking Echo

Here is an example of a View that invokes the Echo FPC:

```go
// Echo models the parameters to be used to invoke the Echo FPC
type Echo struct {
	// Function to invoke
	Function string
	// Args to pass to the function
	Args []string
}

// EchoView models a View that invokes the Echo FPC
type EchoView struct {
	*Echo
}

func (e *EchoView) Call(context view.Context) (interface{}, error) {
	// Invoke the `echo` chaincode deployed on the default channel of the default Fbairc network
	res, err := fpc.GetDefaultChannel(context).Chaincode(
		"echo",
	).Invoke(
		e.Function, fpc.StringsToArgs(e.Args)...,
	).Call()
	assert.NoError(err, "failed invoking echo")
	assert.Equal(e.Function, string(res))

	return res, nil
}

type EchoViewFactory struct{}

func (l *EchoViewFactory) NewView(in []byte) (view.View, error) {
	f := &EchoView{}
	assert.NoError(json.Unmarshal(in, &f.Echo))
	return f, nil
}
```

### Topology

It is very simple to add an FPC to a Fabric topology and have it ready to be used for our integration tests.
The important thing is to have already prepared the docker image with the FPC one wants to deploy.
On how to build an FPC, please, refer to the [`Fabric Private Chaincode documentation`](https://github.com/hyperledger/fabric-private-chaincode#build-fabric-private-chaincode). 
In our case, we will use the following docker image `ghcr.io/mbrandenburger/fpc/fpc-echo:main`.

Here is the topology we use in this case, it is self-explanatory: 

```go
func Topology() []api.Topology {
	// Create an empty fabric topology
	fabricTopology := fabric.NewDefaultTopology()
	// Add two organizations
	fabricTopology.AddOrganizationsByName("Org1", "Org2")
	// Add an FPC by passing chaincode's id and docker image
	fabricTopology.AddFPC("echo", "fpc/fpc-echo")

	// Create an empty FSC topology
	fscTopology := fsc.NewTopology()

	// Alice
	alice := fscTopology.AddNodeByName("alice")
	alice.AddOptions(fabric.WithOrganization("Org2"))
	alice.RegisterViewFactory("ListProvisionedEnclaves", &views.ListProvisionedEnclavesViewFactory{})
	alice.RegisterViewFactory("Echo", &views.EchoViewFactory{})

	// Bob
	bob := fscTopology.AddNodeByName("bob")
	bob.AddOptions(fabric.WithOrganization("Org2"))
	bob.RegisterViewFactory("ListProvisionedEnclaves", &views.ListProvisionedEnclavesViewFactory{})
	bob.RegisterViewFactory("Echo", &views.EchoViewFactory{})

	return []api.Topology{fabricTopology, fscTopology}
}
```