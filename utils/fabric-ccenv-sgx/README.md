# Fabric ccenv with sgx support

Thanks tozd for providing this! https://github.com/tozd/docker-sgx

First do this ...

    $ git clone https://github.com/tozd/docker-sgx.git
    $ cd docker-sgx
    $ ln -s ubuntu-xenial.dockerfile Dockerfile

Next edit ``ubuntu-xenial.dockerfile`` and change the first line (FROM) to the
fabric chaincode environment image.

    FROM hyperledger/fabric-ccenv:latest
    
Also add ``libsystemd-dev`` to the list of packages to install in line 6.

Finally go back to ``fabric-ccenv-sgx`` root directory and just run `make`.

Now you should see ``hyperledger/fabric-ccenv-sgx`` in the list of docker images.

    $ docker images


