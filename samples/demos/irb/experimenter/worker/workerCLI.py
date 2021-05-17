from torchvision import models
import torch
from torch.autograd import Variable
import torch.nn as nn
from train_model import LogisticRegression

import pdo.common.crypto as crypto
import server
import logging
import sys

logging.basicConfig()
logger = logging.getLogger()

logger.setLevel(logging.DEBUG)

import keys

if keys.InitializeKeys() == False:
    logger.error("error initializing keys")
    sys.exit(-1)    

#COMMENT server and exit lines to run original experiment
server.RunWorkerServer()
exit -1



model = LogisticRegression()
model.load_state_dict(torch.load("diagnoss2.pt"))
model.eval()

data0 = [[37.7, 0, 0, 1, 1, 0]]

# the label of data is zero
# for data1 the correct label is one, 1 is disease, 0 is not disease
data1 = [[41.5,0,1,1,0,1]]
input = Variable(torch.tensor(data1, dtype = torch.float32))
prediction = model(input).data.numpy()[:, 0]

# Print out the prediction of probability in precetage having the disease
with open("result.txt", "w") as outfile:
    outfile.write(str(prediction[0] * 100) + "%")

