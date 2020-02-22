/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

const state = {
  inProgress: false
};

const getters = {};

const actions = {
  submit({ commit, state }) {
    return new Promise(resolve => {
      commit("setInProgress", true);
      let id = setInterval(() => {
        if (!state.inProgress) {
          clearInterval(id);
          resolve();
        }
      }, 100);
    });
  },

  finish({ commit }) {
    commit("setInProgress", false);
  },

  clear: ({ commit }) => {
    commit("setInProgress", false);
  }
};

const mutations = {
  setInProgress(state, p) {
    state.inProgress = p;
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
