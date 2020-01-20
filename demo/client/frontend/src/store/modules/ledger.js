/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import LedgerAPI from "@/api/ledger";

const state = {
  transactions: [],
  state: []
};

const getters = {
  hasTransactions: state => {
    return state.transactions.length > 0;
  }
};

const actions = {
  fetchTransactions: ({ commit }) => {
    return LedgerAPI.getLedger()
      .then(response => response.data)
      .then(transactions => {
        commit("setTransactions", transactions.reverse());
      });
  },

  fetchState: ({ commit }) => {
    return LedgerAPI.getState()
      .then(response => response.data)
      .then(data => {
        let state = Object.entries(data).map(([key, value]) => {
          value = atob(value);
          try {
            value = JSON.stringify(JSON.parse(value), null, "\t");
          } catch (e) {
            // value is not a json?
          }
          return { key: key, value: value };
        });

        commit("setState", state);
      });
  },

  updateStateItem: ({ commit }, item) => {
    return LedgerAPI.updateState(item.key, item.value).then(() =>
      commit("setStateItem", item)
    );
  },

  deleteStateItem: ({ commit }, item) => {
    return LedgerAPI.deleteState(item.key).then(() =>
      commit("deleteStateItem", item)
    );
  },

  newTransactionEvent: ({ commit }, event) => {
    return new Promise((resolve, reject) => {
      try {
        let tx = JSON.parse(event.data);
        commit("appendTransaction", tx);
        resolve();
      } catch (err) {
        reject(err);
      }
    });
  },

  clear: ({ commit }) => {
    commit("clearTransactions");
    commit("clearState");
  }
};

const mutations = {
  clearTransactions: state => {
    state.transactions = [];
  },

  clearState: that => {
    that.state = [];
  },

  setTransactions: (that, transactions) => {
    that.transactions = transactions;
  },

  setState: (that, newState) => {
    that.state = newState;
  },

  deleteStateItem: (that, item) => {
    const index = state.state.indexOf(item);
    that.state.splice(index, 1);
  },

  setStateItem: (that, item) => {
    const index = that.state.findIndex(el => el.key === item.key);
    if (index > -1) {
      // get the index and replace
      Object.assign(that.state[index], item);
    } else {
      // just append to state
      that.state.push(item);
    }
  },

  appendTransaction: (that, transaction) => {
    that.transactions.unshift(transaction);
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
