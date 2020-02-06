/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import api from "./api";

export default {
  buildPayload(key, value) {
    return {
      key: key,
      value: value
    };
  },

  getLedger() {
    return api.get(`/ledger`);
  },

  getState() {
    return api.get(`/state`);
  },

  deleteState(key) {
    return api.delete(`/state/` + btoa(key));
  },

  updateState(key, value) {
    return api.post(`/state/` + btoa(key), value);
  }
};
