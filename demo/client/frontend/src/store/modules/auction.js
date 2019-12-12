/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import axios from "axios";
import auction from "@/api/auction";

const state = {
  id: "",
  name: "",
  activityRequirementPercentage: 0,
  clockPriceIncrementPercentage: 0,
  territories: [],
  bidders: [],
  initialEligibilities: [],
  owner: "",
  state: "clock",
  clockRound: 1
};

const getters = {};

const actions = {
  NEW_AUCTION({ commit }, auctionData) {
    auction
      .createAuction(auctionData)
      .then(resp => {
        const status = resp.data.status;
        if (status.rc !== 0) {
          console.log("error: " + status.msg);
        } else {
          auctionData.id = resp.data.response.id;
          commit("SET_AUCTION", auctionData);
        }
      })
      .catch(err => console.log(err));
  },
  LOAD_AUCTION({ commit }, auction_id) {
    axios
      .all([
        auction.getAuctionDetails(auction_id),
        auction.getAuctionStatus(auction_id)
      ])
      .then(
        axios.spread(function(auctionDetails, auctionStatus) {
          // some error checking
          if (auctionDetails.data.status.rc !== 0) {
            console.log("error" + auctionDetails.data.status.msg);
            return;
          }

          if (auctionStatus.data.status.rc !== 0) {
            console.log("error" + auctionStatus.data.status.msg);
            return;
          }

          // only when chaincode returned something
          auctionDetails.data.response.id = auction_id.auctionId;
          commit("SET_AUCTION", auctionDetails.data.response);
          commit("SET_STATUS", auctionStatus.data.response);
        })
      )
      .catch(err => console.log(err));
  },
  END_ROUND({ commit }, auction_id) {
    auction
      .endRound(auction_id)
      .then(resp => {
        const status = resp.data.status;
        if (status.rc !== 0) {
          console.log("error: " + status.msg);
        } else {
          commit("SET_ROUND_ACTIVE", false);
        }
      })
      .catch(err => console.log(err));
  },
  START_NEXT_ROUND({ commit }, auction_id) {
    auction
      .startNextRound(auction_id)
      .then(resp => {
        const status = resp.data.status;
        if (status.rc !== 0) {
          console.log("error: " + status.msg);
        } else {
          commit("SET_ROUND_ACTIVE", true);
          // TODO get auction status
        }
      })
      .catch(err => console.log(err));
  },
  UPDATE_AUCTION_STATE({ commit }, state) {
    commit("SET_AUCTION_STATE", state);
  },
  UPDATE_CLOCK_ROUND({ commit }, round) {
    commit("SET_CLOCK_ROUND", round);
  }
};

const mutations = {
  SET_AUCTION: (state, auction) => {
    state.id = auction.id;
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
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
