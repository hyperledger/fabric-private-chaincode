# Fabric ccenv with sgx support

Thanks tozd for providing this

First do this ...

    git clone https://github.com/tozd/docker-sgx.git
    cd docker-sgx
    ln -s ubuntu-xenial.dockerfile Dockerfile

Next edit `ubuntu-xenial.dockerfile` and change the firs line (FROM) to
the fabric-ccenv image

    FROM hyperledger/fabric-ccenv:latest

Finally go back to fabric-ccenv-sgx root directory and just run `make`


