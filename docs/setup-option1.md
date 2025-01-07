# Setup your Development Environment (Option 1)

In this document we explain how to setup your FPC development environment using the FPC dev docker container. We refer to this setup as `Option 1`. An alternative setup without docker you can find in [Option 2](setup-option2.md).


In this section we explain how to set up a Docker-based development environment that allows you to develop and test FPC chaincode.
The docker images come with all necessary software dependencies and allow you a quick start.
We recommend to set privileges to manage docker as a non-root user. See the
official docker [documentation](https://docs.docker.com/install/linux/linux-postinstall/)
for more details.

First make sure your host has
* Docker v23.0 (or higher).
  It also should use `/var/run/docker.sock` as socket to interact with the daemon (or you
  will have to override in `$FPC_PATH/config.override.mk` the default definition in make of `DOCKER_DAEMON_SOCKET`)
* GNU make

Once you have cloned the repository, you can either use the pre-built images or you can manually build them. After that you will start the development container.

## Pull docker images
To pull the docker image execute the following:
```bash
make -C $FPC_PATH/utils/docker pull pull-dev 
```

## Manually build docker images
In order to build the development image manually you can use the following commands. Note that this process may take some time.
```bash
make -C $FPC_PATH/utils/docker build build-dev 
```

## Start the dev container
Next we will open a shell inside the FPC development container, with environment variables like `$FPC_PATH` appropriately defined and all dependencies like the Intel SGX SDK, ready to build and run FPC.
Continue with the following command:

```bash
make -C $FPC_PATH/utils/docker run-dev
```

Note that by default the dev container mounts your local cloned FPC project as a volume to `/project/src/github.com/hyperledger/fabric-private-chaincode` within the docker container.
This allows you to edit the content of the repository using your favorite editor in your system and the changes inside the docker container. Additionally, you are also not loosing changes inside the container when you reboot or the container gets stopped for other reasons.

A few more notes:
* We use Ubuntu 22.04 by default.
  To build also docker images with a different version of Ubuntu, add the following to `$FPC_PATH/config.override.mk`.
  ```bash
  DOCKER_BUILD_OPTS=--build-arg UBUNTU_VERSION=18.04 --build-arg UBUNTU_NAME=bionic
  ```
* If you run behind a proxy, you will have to configure the proxy,
  e.g., for docker (`~/.docker/config.json`) and load the configuration inside the dev container by setting `DOCKER_DEV_RUN_OPTS += -v "$HOME/.docker":"/root/.docker"` in `$FPC_PATH/config.override.mk`.
  See [Working from behind a proxy](troubleshooting.md#working-from-behind-a-proxy) below for more information. Also note that with newer docker versions (i.e., docker desktop), the docker socket is located on the host in `~/.docker/`. This may cause issues when using docker inside the FPC dev container as the docker client is not able to access the docker socket at the path of the host system. You may try to switch the docker context to use `/var/run/docker.sock`. We do not recommend this approach and happy for suggestions.
* If your local host is SGX enabled, i.e., there is a device `/dev/sgx/enclave` or
  `/dev/isgx` and your PSW daemon listens to `/var/run/aesmd`, then the docker image will be sgx-enabled and your settings from `./config/ias` will be used. You will have to manually set `SGX_MODE=HW` before building anything to use HW mode.
* If you want additional apt packages to be automatically added to your
  container images, you can do so by modifying `$FPC_PATH/config.override.mk` file in the fabric-private-chaincode directory.
  In that file, define
  `DOCKER_BASE_RT_IMAGE_APT_ADD_PKGS`,
  `DOCKER_BASE_DEV_IMAGE_APT_ADD_PKGS'`and/or
  `DOCKER_DEV_IMAGE_APT_ADD_PKGS` with a list of packages you want to be added to you
  all images,
  all images where fabric/fpc is built from source and
  the dev(eloper) container, respectively.
  They will then be automatically added to the docker image.
* Due to the way the peer's port for chaincode connection is managed,
  you will be able to run only a single FPC development container on a
  particular host.
* For support for Apple Mac (M1 or newer) see the [Troubleshooting](../README.md#troubleshooting) section.

Now you are ready to start development *within* the container. Continue with building FPC as described in the [Build Fabric Private Chaincode
](../README.md#build-fabric-private-chaincode) Section and then write [your first Private Chaincode](../README.md#your-first-private-chaincode).
