/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import api from "@/api/api";

const state = {
  status: "",
  name: "",
  role: "",
  avatars: ""
};

const getters = {
  isLoggedIn: state => !!state.name,
  authStatus: state => state.status
};

const actions = {
  login({ commit }, user) {
    return new Promise(resolve => {
      commit("AUTH_SUCCESS", user);
      api.defaults.headers.common["x-user"] = user.id;
      resolve();
    });
  },

  logout({ commit }) {
    return new Promise(resolve => {
      commit("LOGOUT");
      delete api.defaults.headers.common["x-user"];
      resolve();
    });
  }
};

const mutations = {
  AUTH_SUCCESS(state, user) {
    state.status = "success";
    state.name = user.id;
    state.role = user.approle;
    state.avatar = user.avatar;
  },
  LOGOUT(state) {
    state.status = "";
    state.name = "";
    state.role = "";
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
