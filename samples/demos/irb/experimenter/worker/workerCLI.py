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

server.RunWorkerServer()

sys.exit(-1)

