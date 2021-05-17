# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import redis
import logging

logger = logging.getLogger()

class StorageClient:
	def __init__(self, host='localhost', port=6379, db=0):
	    self.host = host
	    self.port = port
	    self.db = db

	    self.r = redis.Redis(self.host, self.port, self.db)

	def set(self, key, value):
	    try:
	        self.r.set(key, value)
	    except Exception as e:
	        logger.error("Error set: {}".format(str(e)))
	        return False
	    return True

	def get(self, key):
	    try:
	        value = self.r.get(key)
	    except Exception as e:
	        logger.error("Error get: {}".format(str(e)))
	        return None, False
	    return value, True

