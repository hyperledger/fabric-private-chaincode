<!---
Licensed under Creative Commons Attribution 4.0 International License
https://creativecommons.org/licenses/by/4.0/
--->
# Chaincode wrapper for Go Chaincode (ecc_go)

wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add
add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v0.3.4/ego_0.3.4_amd64.deb
apt install ./ego_0.3.4_amd64.deb build-essential libssl-dev



cd $FPC_PATH/ecc_go
export CC_NAME=auction
make docker-go-ecc


wget -qO- https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | apt-key add
add-apt-repository "deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu `lsb_release -cs` main"
wget https://github.com/edgelesssys/ego/releases/download/v0.4.0/ego_0.4.0_amd64.deb
apt install ./ego_0.4.0_amd64.deb build-essential libssl-dev



# Kill

docker kill $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)
docker rm $(docker ps -a -q --filter ancestor=fpc/ercc --filter ancestor=fpc/fpc-auction-go)