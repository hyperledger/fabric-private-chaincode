/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import api from "./api";

export default {
  start() {
    return api.get(`/demo/start`);
  }
};
