# Copyright 2021 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

import redis
import logging
from os import environ

logger = logging.getLogger()

class StorageClient:
	def __init__(self):
	    self.host = environ.get('REDIS_HOST', 'localhost')
	    self.port = environ.get('REDIS_PORT', '6379')
	    self.db = '0'

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

