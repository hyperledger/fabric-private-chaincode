# Fabric Private Chaincode goes K8s

This tutorial demonstrates how FPC can be used in a Kubernetes-based Fabric Network. We show how to build and deploy
FPC Chaincode and interact with it using the FPC Client SDK. We use Minikube as a playground for Kubernetes and thereby
allow chaincode developer to setup a local test environment for developing their applications.
Additionally, this tutorial also targets Fabric Operators and illustrates all the necessary steps to deploy applications
using FPC. For the sake of simplicity, this tutorial focuses on the deployment of a single FPC Chaincode, but it should
be easy to extend the provided scripts to also deploy multiple FPC Chaincodes.

Many steps of this tutorial can also be invoked by using `just`. See official [installation](https://github.com/casey/just#installation) documentation. For Mac install via `brew install just`.

## TODOS

- Get this network running with k8s
  -[x] sim mode
  -[ ] hw
- Run this on a k8s cluster
  -[x] minikube
  -[ ] real cluster

## Prepare FPC deployment components

FPC requires a special docker container to execute a FPC chaincode, similar to Fabric's `ccenv` container image but with additional support for Intel SGX.  
You can pull the FPC chaincode environment image (`fabric-private-chaincode-ccenv`) from our Github repository or build them manually as follows:

```bash
# pulls the fabric-private-chaincode-ccenv image from github
make -C $FPC_PATH/utils/docker pull

# builds fabric-private-chaincode-ccenv image from scratch
make -C $FPC_PATH/utils/docker build
```

Then we build the FPC components including the FPC Enclave Registry docker image.
```bash
make -C $FPC_PATH build
make -C $FPC_PATH/ercc docker
```

### Build your FPC Chaincode

For this demo we use the FPC Echo Chaincode in `$FPC_PATH/samples/chaincode/echo`.
Alternatively, you can also use another example or your own FPC Chaincode.
We build the chaincode and package the enclave binary in a FPC Chaincode docker image using the following command:

```bash
# TODO SETUP and test with SGX HW mode
export TEST_CC_PATH=$FPC_PATH/samples/chaincode/echo

make -C $TEST_CC_PATH
make -C $FPC_PATH/ecc DOCKER_IMAGE=fpc/fpccc DOCKER_ENCLAVE_SO_PATH=$TEST_CC_PATH/_build/lib all docker
```

Note that this command results in a docker images with the name `fpc/fpccc` containing the FPC Echo Chaincode.
We use this generic image name for simplicity here, so a developer can use easily "override" this image with another
FPC Chaincode image. For non-development purpose we recommend a more descriptive image name.


### Writing your FPC Client App

In order to communicate with the FPC Chaincode we use an app written in Go using our FPC Client SDK.
In this tutorial we will use a simple CLI app in `$FPC_PATH/samples/application/simple-cli-go/`.

```bash
cd $FPC_PATH/samples/application/simple-cli-go/
make build docker
```

Note that this commands create a docker image with the name `fpc/fpcclient`.

## Minikube on Mac

This tutorial is heavily inspired by the article [How to implement Hyperledger Fabric External Chaincodes within a Kubernetes cluster](https://medium.com/swlh/how-to-implement-hyperledger-fabric-external-chaincodes-within-a-kubernetes-cluster-fd01d7544523) by Pau Aragonès Sabaté. Thank you!

### Start minikube

First start minikube and mount our local `$FPC_PATH` into minikube.
```bash
cd $FPC_PATH
minikube start --mount --mount-string="$(pwd):/fpc"

# Let's start he k8s dashboard
minikube dashboard
```

#### Push FPC images to minikube registry

First let's go the demo folder

```bash
cd $FPC_PATH/samples/deployment/k8s
```

In order to use our docker images for the Enclave Registry, the FPC Chaincode, and the FPC Client App, we need to
make these images available inside the k8s environment.
With minikube we can easily do that by calling the following commands:

```bash
minikube cache add fpc/ercc:latest
minikube cache add fpc/fpccc:latest
minikube cache add fpc/fpcclient:latest
```

You can double check if the images are available. Please also carefully check if the image ID corresponds to the image you want to use.
```bash
minikube ssh
docker images | grep fpc
```

When you update docker images, you need to tell minikube to use the updated images.
```bash
minikube cache reload
```

Note that sometimes `cache reload` appears to not work properly. In some cases it may be better to build the FPC
docker images within the docker context of minikube using `eval $(minikube docker-env)`.

### Prepare K8s network

In the next step we prepare the crypto material used in our network. This requires the `cryptogen` and the `configtxgen`
tools. We assume that these tools are available in `$FABRIC_PATH/build/bin/`. 
If you have installed them somewhere else on your system, please set `FABIC_BIN_DIR` accordingly.

For instance, you can download the binaries and use them by following the commands:
```bash
cd $FPC_PATH/samples/deployment/k8s
curl -sSL https://bit.ly/2ysbOFE | bash -s -- 2.3.3 1.4.9 -d -s
export FABRIC_BIN_DIR=$(pwd)/bin
```

Use the `generate.sh` script to create the crypto material for the peer organizations and the orderer. Moreover, this
script create the genesis block, the client connection profiles, and packages the FPC Chaincode and the FPC Enclave Registry.
```bash
cd $FPC_PATH/samples/deployment/k8s
./generate.sh
````

Prepare k8s environment:
```bash
kubectl create ns hyperledger
kubectl create configmap peer-config --from-file=core.yaml -n hyperledger
kubectl create configmap chaincode-config --from-env-file=packages/chaincode-config.properties -n hyperledger
```

Deploy the Fabric network
```bash
kubectl create -f orderer-service/
kubectl create -f org1/
kubectl create -f org2/
kubectl create -f org3/
```

Alternatively, you can also call `just generate up`.

### Setting up our channel

In the next steps we need three terminals, each dedicated for a particular organization, to enter the CLI container.

Show all pods
```bash
# list all pods
kubectl get pods -n hyperledger
```

Org1
```bash
# copy cli_org1 pod name from above and enter container
kubectl exec -it cli_org1_pod_name -n hyperledger -- bash

peer channel create -o orderer0:7050 -c mychannel -f ./channel-artifacts/channel.tx --tls true --cafile $ORDERER_CA
peer channel join -b mychannel.block
```

Org2 and Org3
```bash
# copy cli_org1 pod name from above and enter container
kubectl exec -it cli_orgX_pod_name -n hyperledger -- bash

peer channel fetch 0 mychannel.block -c mychannel -o orderer0:7050 --tls --cafile $ORDERER_CA
peer channel join -b mychannel.block
peer channel list
```

Alternatively, you can use `just run cli-org1` to login into the cli instance of org1.

#### Set anchor peers

In order to allow the peers to discover peers of another organization we need to setup an anchor peer for each org.
Since we only have a single peer per org, we set each peer as an anchor peer for their org.

As this task is somewhat tedious we provide a little helper script.

Execute the following command for each org in corresponding CLI container.
```bash
./scripts/setAnchorPeer.sh
```

For more details on anchor peer setup and channel configuration updates please see the official Hyperledger Fabric
[documentation](https://hyperledger-fabric.readthedocs.io/en/latest/config_update.html). 

### Install FPC Enclave Registry and FPC ECC

Now it is time to install FPC components on our network.
In particular, we need to install the FPC Enclave Registry (ERCC) and the actual chaincode.
With FPC, there are two options available to run these components based on the external launcher feature of Fabric.
We can run these components as normal Chaincode without docker or as chaincode as a Server (External chaincode).

#### Normal Chaincode without docker
The first option requires our custom FPC peer images, which provide the necessary software package to run Intel SGX applications. 

#### Chaincode as a Server (External chaincode)
The second option allows us to use standard Fabric peer images but requires the deployer to setup and run the chaincodes as a server.
Thus, the responsibility to maintain the chaincodes is in the hand of the network operator.

In this tutorial we show how to setup our FPC components using the Chaincode as a Server method.
Note that the `core.yaml` we have already loaded in the configmap contains the necessary configuration details to use the external builder scripts provided in the FPC repository. 

#### Install Enclave Registry and FPC Chaincode

For each Org:
```bash
# copy cli_orgX pod name from above and enter container
kubectl exec -it cli_orgX_pod_name bash -n hyperledge

# for all peers of orgX we install ercc and fpccc
export ERCC=ercc-peer0-$ORG; export FPCCC=fpccc-peer0-$ORG

peer lifecycle chaincode install packages/$ERCC.tgz
peer lifecycle chaincode install packages/$FPCCC.tgz
peer lifecycle chaincode queryinstalled

export ERCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep ercc | awk '{print $3}' | sed 's/.$//')
export FPCCC_PKG_ID=$(peer lifecycle chaincode queryinstalled | grep fpccc | awk '{print $3}' | sed 's/.$//')

echo "$ERCC=$ERCC_PKG_ID" >> packages/chaincode-config.properties
echo "$FPCCC=$FPCCC_PKG_ID" >> packages/chaincode-config.properties
```

Alternatively, you can also run `./scripts/installCC.sh` inside the cli_orgX_pod_name.

#### Start chaincode container

Now, we have installed our FPC Enclave Registry and the FPC Chaincode on each peer. Since we are using the chaincode
as s server deployment model, it is now time to start the chaincodes.

Run the following commands from your host terminal.

```bash
kubectl create configmap chaincode-config --from-env-file=packages/chaincode-config.properties -n hyperledger --dry-run=client -o yaml | kubectl apply -f -
kubectl create -f chaincode/ercc/
kubectl create -f chaincode/fpccc/
```

Alternatively, you can also call `just chaincode`.

#### Approve

Next, we complete chaincode installation process by approving and committing the chaincode definitions.

For each Org:
```bash
# copy cli_orgX pod name from above and enter container
kubectl exec -it cli_orgX_pod_name bash -n hyperledge

# make sure you still have ERCC_PKG_ID, FPCCC_PKG_ID, and FPC_MRENCLAVE defined in your env (see above)

# approve enclave registry
peer lifecycle chaincode approveformyorg --channelID mychannel --name ercc --version 1.0 --package-id $ERCC_PKG_ID --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA

# approve FPC Chaincode
peer lifecycle chaincode approveformyorg --channelID mychannel --name fpccc --version $FPC_MRENCLAVE --package-id $FPCCC_PKG_ID --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
```

Alternatively, you can also run `./scripts/approve.sh` inside the cli_orgX_pod_name.

#### Commit

Pick some org that proceeds with committing chaincode.

```bash
# commit enclave registry
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name ercc --version 1.0 --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
peer lifecycle chaincode commit -o orderer0:7050 --channelID mychannel --name ercc --version 1.0 --sequence 1 --tls true --cafile $ORDERER_CA --peerAddresses peer0-org1:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/ca.crt --peerAddresses peer0-org2:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/peers/peer0-org2/tls/ca.crt

# commit FPC Chaincode
peer lifecycle chaincode checkcommitreadiness --channelID mychannel --name fpccc --version  $FPC_MRENCLAVE --sequence 1 -o orderer0:7050 --tls --cafile $ORDERER_CA
peer lifecycle chaincode commit -o orderer0:7050 --channelID mychannel --name fpccc --version  $FPC_MRENCLAVE --sequence 1 --tls true --cafile $ORDERER_CA --peerAddresses peer0-org1:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/ca.crt --peerAddresses peer0-org2:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2/peers/peer0-org2/tls/ca.crt
```

### Interact with our FPC Chaincode

```bash
# and join into the container
kubectl exec -it fpcclient_orgX_pod_name bash -n hyperledge

# init our enclave
fpcclient init peer0-org1

# interact with the FPC Chaincode
fpcclient invoke eat apple
fpcclient query say apple 
```

Alternatively, you can also run `just run fpcclient-org1` to enter the container

### Shutdown cluster
```bash
kubectl delete --all all --namespace=hyperledger
```

Alternatively, you can also call `just down`.


## Troubleshooting

```bash
discover --configFile conf.yaml --peerTLSCA /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/ca.crt --tlsKey /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/server.key --tlsCert /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/peers/peer0-org1/tls/server.crt --userKey /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/users/User1@org1/msp/keystore/priv_sk --userCert /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1/users/User1@org1/msp/signcerts/User1@org1-cert.pem --MSP org1MSP saveConfig
discover --configFile conf.yaml peers --channel mychannel --server peer0-org1:7051
```


## FAQ

TODO find answers ....

- the fabric peers are already fpc-enabled. so what if you have an existing fabric network? do you add a peer? do you shutdown a peer, modify its configuration and bring it up again?

- the discovery of the peer is up to fabric, but the discovery of the enclaves is up to FPC. Is the endpoint that FPC stores in ERCC sufficient?

- the current implementation hardcoded an fpc chaincode with the fpccc name, which makes impossible to run a separate one. While this was made to address (1), what's required to avoid this name-dependency?

- other questions related to chaincode updates, peer restart, etc. are left for later

- What happens in the case of a incremental FPC deployment?
  - Would it be possible to convert a Fabric peer into an FPC-enabled peer? 
  - How about adding an FPC-enabled peer?