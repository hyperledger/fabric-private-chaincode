# Copyright IBM Corp. All Rights Reserved.
# Copyright 2020 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM python:3.8

# set the working directory in the container
WORKDIR /code

RUN apt-get update -q \
    && apt-get install -y -q \
    libnss-mdns \
    libnss-myhostname \
    lsb-release

RUN pip install torch==1.8.1+cpu torchvision==0.9.1+cpu torchaudio==0.8.1 -f https://download.pytorch.org/whl/torch_stable.html
RUN pip install pillow twisted
RUN pip install redis

# copy experiment code
COPY worker/ .

# expose server port
EXPOSE 5000

CMD ["python3", "workerCLI.py"]
