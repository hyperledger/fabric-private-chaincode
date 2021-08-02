# FPC docker 

## Use cases

The docker images provided in this repository target different use cases:

1) *Used as FPC dev environment*
   
   Provides a complete development environment including Intel SGX compiler, etc... 
   Useful to develop and test FPC locally without installing FPC dependencies on local dev machine. 
   This refers to preferred option to getting started with FPC.

1) *Used as CI environment*
   
   Based on FPC dev environment and used during Github CI. 

1) *Used as runtime environment for FPC components (FPC chaincode and ERCC)*
   
   Required to package, deploy, and run FPC chaincode. 


## Docker Images

FPC comes with the following docker images to build FPC components and the development environment.
All images start with `hyperledger/fabric-private-chaincode-` prefix.

* `base-rt`: Base image for FPC. Includes all runtime dependencies including SGX runtime services.
* `ccenv`: Chaincode environment image for FPC chaincodes based on `base-rt`.
* `base-dev`: Base development image. Includes all build tools for FPC including protobuf, SGX SSL, SGX compiler, ...
* `dev`: Development image based on `base-dev`. Add additional user-defined tools and development dependencies.

These images can be build manually (see [Building images](#building-images) section) or pulled from `ghcr.io/hyperledger/fabric-private-chaincode` (see [Pulling images](#pulling-images) section).


## Building images

* `make build`: creates base-rt, ccenv
  
* `make build-dev`: creates base-rt, base-dev
  

## Pulling images

* `make pull`: pulls ccenv

* `make pull-dev`: pulls base-dev 

Note that base-rt not pulled as it is an intermediate image included already as layer in `ccenv` and `base-dev`.

## Running FPC dev environment

* `make run-dev`: creates dev (if not exist) and runs it. Does not create base-dev and base-rt, returns an error if not exists.


## Usage

### Start docker-based FPC dev environment

Option 1) Pull images and start dev container

```bash
cd utils/docker
make pull-dev
make run-dev
## continue inside docker
```

Option 2) Build images from scratch and start dev container
```bash
cd utils/docker
make build-dev
make run-dev
## continue inside docker
```


### CI

If no changes in `utils/docker`
```bash
cd utils/docker
make pull
make run-dev
## continue inside docker
```

otherwise
```bash
cd utils/docker
make build-dev
make run-dev
## continue inside docker
```


### Build docker images for FPC runtime environment

Pull images
```bash
cd utils/docker
make pull
```

or build them
```bash
cd utils/docker
make build
```


## Publishing

We publish images for every new release (docker tag = `${FPC_VERSION}`) and on PR merged (docker tag = main) through CI.

* `make publish`: pushing ccenv and base-dev to Github docker registry 

When running the `publish` target manually (not through CI), docker login on `ghcr.io` is needed.
See [documentation](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry).

Use your github personal access token (PAT) to login.
```bash
export CR_PAT=YOUR_TOKEN
echo $CR_PAT | docker login ghcr.io -u USERNAME --password-stdin
```

Note that we only publish Ubuntu 20.04 LTS based images.

### CI/CD
TODO describe automated publishing

