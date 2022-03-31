<!---
Copyright (c) Siemens AG, 2022
SPDX-License-Identifier: Apache-2.0
--->
# How to use an Azure Confidential Computing Instance to run FPC in HW mode
Date of writing 14.03.2022.
This guide is based on an article by [Koshi Ikegawa](https://qiita-com.translate.goog/ikegawa-koshi/items/8cf1fef1004fc16beb15?_x_tr_sl=ja&_x_tr_tl=en&_x_tr_hl=de&_x_tr_pto=wapp#fpc%E3%81%AE%E3%83%93%E3%83%AB%E3%83%89).

  - [Creating the Confidential Computing Instance on Azure.](#creating-the-confidential-computing-instance-on-azure)
  - [Prerequisites](#prerequisites)
  - [Registering for the SGX Attestation Service Utilizing EPID](#registering-for-the-sgx-attestation-service-utilizing-epid)
  - [Setting up the FPC development environment](#setting-up-the-fpc-development-environment)
  - [Final preparation of the dev-container](#final-preparation-of-the-dev-container)

## Creating the Confidential Computing Instance on Azure.
Use the [Quick Create Portal](https://docs.microsoft.com/en-us/azure/virtual-machines/linux/quick-create-portal) to create the virtual machine.
Use the following parameters:
* OS = *Linux (ubuntu 20.04)*
* Size = Standard DC2s v2 (2 vcpus, 8 GiB memory)

## Prerequisites
Connect to your instance using ssh and execute the following commands:
```bash
# update machine
sudo -i
apt-get update
apt-get upgrade
apt install -y docker.io make
# install sgx environment
# 1.add sgx repo + key and then install packages form this repository
echo 'deb [arch=amd64] https://download.01.org/intel-sgx/sgx_repo/ubuntu focal main' | sudo tee /etc/apt/sources.list.d/intel-sgx.list
wget -qO - https://download.01.org/intel-sgx/sgx_repo/ubuntu/intel-sgx-deb.key | sudo apt-key add -
sudo apt-get update
sudo apt -y install libssl-dev libsgx-enclave-common libsgx-enclave-common-dev libsgx-ae-qe3 libsgx-ae-qve libsgx-epid libsgx-launch libsgx-pce-logic libsgx-qe3-logic libsgx-quote-ex libsgx-uae-service libsgx-urts
sudo reboot
```
After reboot is completed connect to your vm again.
```bash
# check if asemd service is running
sudo systemctl status aesmd.service
```
```bash
sudo usermod -aG docker $(whoami)
sudo reboot
``` 
After reboot is completed connect to your vm again.
```bash
export GO_PATH=$HOME/go
export FPC_PATH=$GOPATH/src/github.com/hyperledger/fabric-private-chaincode 
git clone --recursive https://github.com/hyperledger/fabric-private-chaincode.git $FPC_PATH
```

## Registering for the SGX Attestation Service Utilizing EPID
* Get a account [here](https://www.intel.com/content/www/us/en/forms/basic-intel-registration.html)
* Once you are signed in go [here](https://api.portal.trustedservices.intel.com/EPID-attestation)
* Subscribe for development unlikable.
  * You could also use the linkable attestation. To make it work you have to change the contents file **spid_type.txt** from *epid-unlinkable* to *epid-linkable*.
* You will receive a SPID, Primary Key and Secondary Key.
* Use these information to replace the appropriate places in the commands below.

```bash
echo '[YOUR_SPID]' > ${FPC_PATH}/config/ias/spid.txt
echo '[YOUR_PRIMARY_KEY]' > ${FPC_PATH}/config/ias/api_key.txt
echo 'epid-unlinkable' > ${FPC_PATH}/config/ias/spid_type.txt
```

## Setting up the FPC development environment 
There are two methods of setting up the FPC development environment.
The [docker based](../../README.md#option-1-using-the-docker-based-fpc-development-environment) environment which is used here, and the [local development](../../README.md#option-2-setting-up-your-system-to-do-local-development) environment.
Edit the ``config.override.mk``  to set HW mode and ``DOCKER_BUILDKIT=1``.
```bash
vim $FPC_PATH/config.override.mk
```
paste in the following
```bash
export SGX_MODE=HW
export DOCKER_BUILDKIT=1
```
Now we can build the container as follows:
```bash
cd ${FPC_PATH}/utils/docker
make pull-dev
```
**After** make run-dev you will end up in a shell in the dev container, you need to enter **exit**.Otherwise the container stops if you ever accidentally exit the container later.
```
make run-dev
exit
```
Now, you can contine building FPC [here](../../README.md#build-fabric-private-chaincode)
* Start the container: ``docker start fpc-development-main``
* Open a shell in the container: ``docker exec -i -t fpc-development-main bash``
* Stop the container: ``docker stop fpc-development-main``