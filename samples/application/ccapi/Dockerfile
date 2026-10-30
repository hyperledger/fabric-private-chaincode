# Use an official Golang runtime as a parent image
FROM golang:1.27.3-alpine AS build

ENV PATH="${PATH}:/usr/bin/"

RUN apk update 

RUN apk add \
    docker \
    openrc \
    git \ 
    gcc \
    gcompat \
    libc-dev  \
    libc6-compat  \
    libstdc++ && \
    ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2

# The build context is the FPC repository root, as go.mod replaces
# fabric-private-chaincode with ../../../ (see Dockerfile.dockerignore)
WORKDIR /fpc

# Copy the go.mod and go.sum files for dependency management
COPY go.mod go.sum ./
COPY samples/application/ccapi/go.mod samples/application/ccapi/go.sum samples/application/ccapi/

# Install go dependencies
WORKDIR /fpc/samples/application/ccapi
RUN go mod download

# Copy the FPC packages used by ccapi and the ccapi sources
COPY client_sdk /fpc/client_sdk
COPY internal /fpc/internal
COPY samples/application/ccapi .

# Build the Go ccapi
RUN go build -o ccapi

# # Use an official Alpine runtime as a parent image
# FROM alpine:latest

# ENV PATH="${PATH}:/usr/bin/"

# RUN apk update 

# RUN apk add \
#     docker \
#     openrc \
#     git \ 
#     gcc \
#     gcompat \
#     libc-dev  \
#     libc6-compat  \
#     libstdc++ && \
#     ln -s /lib/libc.so.6 /usr/lib/libresolv.so.2

# # Set the working directory to /rest-server
# WORKDIR /rest-server

# Copy the ccapi binary from the build container to the current directory in the Alpine container
RUN ls -l
RUN pwd

RUN cp ./ccapi /usr/bin/ccapi

# Run the ccapi binary
CMD ["ccapi"]