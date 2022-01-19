<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Chaincode wrapper for Go Chaincode (ecc_go)



# Install Ego inside dev environment

```bash
wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add
add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v0.4.0/ego_0.4.0_amd64.deb
apt install ./ego_0.4.0_amd64.deb build-essential libssl-dev
```

Prepare `ccenv-go` image with
```bash
cd $FPC_PATH/utils/docker/
make ccenv-go
```

# Example

```bash
cd $FPC/samples/chaincode/auction-go
make 
```

# Developer notes

## Docker container

```bash
docker tag fpc/ercc:main fpc/ercc:latest
```

## Kill

docker kill $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)
docker rm $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)