/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import auction from "@/api/auction";

const state = {
  submittedBids: []
};

const getters = {};

const actions = {
  submitBid({ commit }, bid) {
    auction
      .submitClockBid(bid)
      .then(response => {
        console.log(response);
        commit("pushBid", bid);
      })
      .catch(err => console.log(err));
  }
};

const mutations = {
  pushBid(state, payload) {
    state.submittedBids.push(payload);
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
