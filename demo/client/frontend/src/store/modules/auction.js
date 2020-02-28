/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import auction from "@/api/auction";
import helpers from "../helpers";

const defaultsState = {
  id: "",
  name: "",
  activityRequirementPercentage: 0,
  clockPriceIncrementPercentage: 0,
  territories: [],
  bidders: [],
  initialEligibilities: [],
  owner: "",
  state: "clock",
  clockRound: 1,
  roundActive: false,
  winner: "",
  results: ""
};

const state = defaultsState;

const getters = {};

const actions = {
  NEW_AUCTION({ commit }, auctionData) {
    return auction
      .createAuction(auctionData)
      .then(resp => helpers.checkStatus(resp.data))
      .then(auction => {
        commit("SET_AUCTION_ID", auction.auctionId);
        commit("SET_AUCTION", auctionData);
      });
  },

  LOAD_AUCTION({ commit }, auction_id) {
    auction_id = { auctionId: auction_id };
    return axios
      .all([
        auction
          .getAuctionDetails(auction_id)
          .then(resp => helpers.checkStatus(resp.data)),
        auction
          .getAuctionStatus(auction_id)
          .then(resp => helpers.checkStatus(resp.data))
      ])
      .then(
        axios.spread(function(auctionDetails, auctionStatus) {
          // only when chaincode returned something
          commit("SET_AUCTION_ID", auction_id.auctionId);
          commit("SET_AUCTION", auctionDetails);
          commit("SET_STATUS", auctionStatus);
        })
      );
  },

  UPDATE_STATUS({ commit }, auction_id) {
    auction_id = { auctionId: auction_id };
    return auction
      .getAuctionStatus(auction_id)
      .then(resp => helpers.checkStatus(resp.data))
      .then(auctionStatus => commit("SET_STATUS", auctionStatus));
  },

  END_ROUND({ commit }, auction_id) {
    auction_id = { auctionId: auction_id };
    return auction
      .endRound(auction_id)
      .then(resp => helpers.checkStatus(resp.data))
      .then(() => commit("SET_ROUND_ACTIVE", false));
  },

  NEXT_ROUND({ commit }, auction_id) {
    auction_id = { auctionId: auction_id };
    return auction
      .startNextRound(auction_id)
      .then(resp => helpers.checkStatus(resp.data))
      .then(winner => commit("SET_WINNER", winner));
  },

  UPDATE_AUCTION_STATE({ commit }, state) {
    commit("SET_AUCTION_STATE", state);
  },

  UPDATE_CLOCK_ROUND({ commit }, round) {
    commit("SET_CLOCK_ROUND", round);
  },
  UPDATE_RESULTS({ commit }, auction_id) {
    auction_id = { auctionId: auction_id };
    return auction
      .getAssignmentResults(auction_id)
      .then(resp => helpers.checkStatus(resp.data))
      .then(data => commit("SET_AUCTION_RESULT", data.result));
  },

  clear: ({ commit }) => {
    commit("clearState");
  }
};

const mutations = {
  SET_AUCTION_ID: (state, auctionId) => {
    state.id = auctionId;
  },

  SET_AUCTION: (state, auction) => {
    state.name = auction.name;
    state.activityRequirementPercentage = auction.activityRequirementPercentage;
    state.clockPriceIncrementPercentage = auction.clockPriceIncrementPercentage;
    state.territories = auction.territories;
    state.bidders = auction.bidders;
    state.initialEligibilities = auction.initialEligibilities;
    state.owner = auction.owner;
  },
  SET_STATUS: (state, status) => {
    state.clockRound = status.clockRound;
    state.state = status.state;
    state.roundActive = status.roundActive;
  },
  SET_ROUND_ACTIVE: (state, active) => {
    state.roundActive = active;
  },
  SET_AUCTION_STATE: (state, auction_state) => {
    state.state = auction_state;
  },
  SET_CLOCK_ROUND: (state, round) => {
    state.clockRound = round;
  },
  SET_WINNER: (state, winner) => {
    state.winner = winner;
  },
  SET_AUCTION_RESULT: (state, auction_result) => {
    state.results = auction_result;
  },
  clearState: that => {
    that.state = that.defaultState;
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
