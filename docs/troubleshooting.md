# Troubleshooting

This document tries to give some help if something is not working as expected. If you experience any others issue and want to share this, please do not hesitate to create a PR and add to this section.

## Docker

Building the project requires docker. We do not recommend to run `sudo make`
to resolve issues with mis-configured docker environments as this also changes your `$GOPATH`. Please see hints on
[docker](#docker) installation above.

The makefiles do not ensure that docker files are always rebuild to
match the latest version of the code in the repo.  If you suspect you
have an issue with outdated docker images, you can run `make clobber
build` which forces a rebuild.  It also ensures that all other
download, build or test artifacts are scrubbed from your repo and might
help overcoming other problems. Be advised that that the rebuild can
take a fair amount of time.

## Working from behind a proxy

The current code should work behind a proxy assuming
  * you have defined the corresponding environment variables (i.e.,
    `http_proxy`, `https_proxy` and, potentially, `no_proxy`) properly, and
  * docker (daemon & client) is properly set up for proxies as
    outlined in the Docker documentation for
    [clients](https://docs.docker.com/network/proxy/) and the
    [daemon](https://docs.docker.com/config/daemon/systemd/#httphttps-proxy).
  * the docker version is correct.
    Otherwise you may run into problems with DNS resolution inside the container.
  * the docker-compose version is correct.
    For example, the docker-compose from Ubuntu 18.04 (docker-compose 1.17)
    is _not_ recent enough to understand `~/.docker/config.json` and related proxy options.

Furthermore, for docker-compose networks to work properly with proxies, the `noProxy`
variable in your `~/.docker/config.json` should at least contain `127.0.0.1,127.0.1.1,localhost,.org1.example.com,.example.com`.

Another problem you might encounter when running the integration tests
insofar that some '0.0.0.0' in `integration/config/core.yaml` used by
clients -- e.g., the peer CLI using the `address: 0.0.0.0:7051` config
as part of the `peer` section -- result in the client being unable
to find the server. The likely error you will see is
 `err: rpc error: code = Unavailable desc = transport is closing`.
In that case, you will have to replace the '0.0.0.0' with a concrete
ip address such as '127.0.0.1'.


## Environment settings

Our build system requires a few variables to be set in your environment. Missing variables may cause `make` to fail.
Below you find a summary of all variables which you should carefully check and add to your environment.

```bash
# Path to your SGX SDK and SGX SSL
export SGX_SDK=/opt/intel/sgxsdk
export SGX_SSL=/opt/intel/sgxssl

# Path to nanopb
export NANOPB_PATH=$HOME/nanopb

# SGX simulation mode
export SGX_MODE=SIM

# SGX simulation mode
export SGX_MODE=HW
```
The file `config.mk` contains various defaults for some of these, but
all can be (re)defined also in an optional file `config.override.mk`.


## Clang-format

Some users may experience problems with clang-format. In particular, the error `command not found: clang-format`
appears even after installing it via `apt-get install clang-format`. See [here](https://askubuntu.com/questions/1034996/vim-clang-format-clang-format-is-not-found)
for how to fix this.

## ERCC setup failures

<!-- TODO: check below, this section is probably outdated? -->

If, e.g., running the integration tests executed when you run `make`,
you get errors of following form:

```
Error: endorsement failure during invoke. response: status:500 message:"Setup failed: Can not register enclave at ercc: Error while retrieving attestation report: IAS returned error: Code 401 Access Denied"
```

In case you run in SGX HW mode, check that your files in `config/ias`
are set properly as explained in [Section Intel Attestation Service
(IAS)](../docs/build-sgx.md#register-with-intel-attestation-service-ias).  Note that if you run
initially in simulation mode and these files do not exist, the build
will create dummy files. In case you switch later to HW mode without
configuring these files correctly for HW mode, this will result in
above error.


## No Raft leader

The following error message sometimes appears when running the integration tests in the `$FPC_PATH/integration` folder.
The output contains the following:
```
got unexpected status: SERVICE_UNAVAILABLE -- no Raft leader
```

Rerunning the tests usually works.
If this error appers during the make step of [building FPC](../README.md#build-fabric-private-chaincode) than uncommenting some integration tests fixes the issue.


## Working with the FPC dev container

To make starting and stopping the dev container more reliable it is advised to use the following commands:
* Start the container and get a shell: `make -C $FPC_PATH/utils/docker run-dev`
* Get another shell inside the dev container: `docker exec -it fpc-development-main /bin/bash`
* Stop the container: `docker stop fpc-development-main`

## Development on Apple Mac (M1 or newer)

For developers using Apple Mac (M1 or newer) we suggest to use the prebuilt FPC dev container.
Add the following configuration to your `config.override.mk`, pull the docker images and start the FPC dev container as described above in [Option 1: Using the Docker-based FPC Development Environment](../README.md#option-1-using-the-docker-based-fpc-development-environment).
Note that SGX is not supported on Apple platforms, and hence, FPC chaincode can only be used in simulation mode.
Alternatively, a cloud-based development environment can be used with SGX HW support, see our tutorial [How to use FPC with Azure Confidential Computing](../samples/deployment/azure/FPC_on_Azure.md).

```Makefile
DOCKER_BUILD_CMD=buildx build
DOCKER_BUILD_OPTS=--platform linux/amd64
DOCKER_DEV_RUN_OPTS=--platform linux/amd64
```
