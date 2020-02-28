/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import auction from "@/api/auction";
import helpers from "../helpers";

const state = {
  submittedBids: []
};

const getters = {};

const actions = {
  submitBid({ commit }, bid) {
    return auction
      .submitClockBid(bid)
      .then(response => helpers.checkStatus(response.data))
      .then(() => commit("pushBid", bid));
  },

  clear: ({ commit }) => {
    commit("clearState");
  }
};

const mutations = {
  pushBid(state, payload) {
    state.submittedBids.push(payload);
  },

  clearState: that => {
    that.submittedBids = [];
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
