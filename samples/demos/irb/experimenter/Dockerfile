# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM python:3.8

# set the working directory in the container
WORKDIR /irb/experimenter/worker

RUN apt-get update -q \
    && apt-get install -y -q \
    libnss-mdns \
    libnss-myhostname \
    lsb-release \
    swig

RUN pip install torch==1.8.1+cpu torchvision==0.9.1+cpu torchaudio==0.8.1 -f https://download.pytorch.org/whl/torch_stable.html
RUN pip install pillow twisted
RUN pip install redis
RUN pip install protobuf
RUN pip install tensorflow
RUN pip install matplotlib
RUN pip install cryptography

## copy pdo crypto patch
#COPY experimenter/pdo_python_setup.py.patch /tmp
#
#RUN cd /tmp \
#    && git clone https://github.com/hyperledger-labs/private-data-objects.git \
#    && cd private-data-objects/python \
#    && git apply /tmp/pdo_python_setup.py.patch \
#    && python3.8 setup.py build_ext

# copy experimenter code
COPY experimenter/worker /irb/experimenter/worker
COPY pkg /irb/pkg
COPY protos /irb/protos

# expose server port
EXPOSE 5000

CMD ["sh", "-c", "PYTHONPATH=../../ python3 workerCLI.py"]