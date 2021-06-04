# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

from torchvision import models
import torch
from torch.autograd import Variable
import torch.nn as nn
from train_model import LogisticRegression

# This function performs one diagnosis and returns a result string and an error string (Go style)
def diagnose(data_bytes):
    data_string = data_bytes.decode()
    data_items = data_string.split(',')
    if len(data_items) != 6:
        return "", "wrong number of parameters: " + str(len(data_items))

    data_items[0] = float(data_items[0])
    data_items[1] = int(data_items[1])
    data_items[2] = int(data_items[2])
    data_items[3] = int(data_items[3])
    data_items[4] = int(data_items[4])
    data_items[5] = int(data_items[5])

    data_list = [data_items]

    model = LogisticRegression()
    model.load_state_dict(torch.load("diagnoss2.pt"))
    model.eval()

    input = Variable(torch.tensor(data_list, dtype = torch.float32))
    prediction = model(input).data.numpy()[:, 0]

    return "Decision: " + str(round(prediction[0] * 100, 5)) + "% of having Nephritis of renal pelvis origin", ""

