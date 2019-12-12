/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import api from "./api";

const invoke = "/cc/invoke";
const query = "/cc/query";

export default {
  buildPayload(tx, args) {
    return {
      tx: tx,
      args: [JSON.stringify(args)]
    };
  },

  ///////////////////////////////////////////////
  // Helpers
  ///////////////////////////////////////////////

  getDefaultAuction() {
    return api.get("/clock_auction/getDefaultAuction");
  },

  ///////////////////////////////////////////////
  // Auctioneer actions
  ///////////////////////////////////////////////

  createAuction(args) {
    return api.post(`${invoke}`, this.buildPayload("createAuction", args));
  },

  getAuctionDetails(args) {
    return api.post(`${query}`, this.buildPayload("getAuctionDetails", args));
  },

  getAuctionStatus(args) {
    return api.post(`${query}`, this.buildPayload("getAuctionStatus", args));
  },

  startNextRound(args) {
    return api.post(`${invoke}`, this.buildPayload("startNextRound", args));
  },

  endRound(args) {
    return api.post(`${invoke}`, this.buildPayload("endRound", args));
  },

  ///////////////////////////////////////////////
  // Bidder actions
  ///////////////////////////////////////////////

  submitClockBid(args) {
    return api.post(`${invoke}`, this.buildPayload("submitClockBid", args));
  },

  getRoundInfo(args) {
    return api.post(`${query}`, this.buildPayload("getRoundInfo", args));
  },

  getBidderRoundResults(args) {
    return api.post(
      `${query}`,
      this.buildPayload("getBidderRoundResults", args)
    );
  },

  getOwnerRoundResults(args) {
    return api.post(
      `${query}`,
      this.buildPayload("getOwnerRoundResults", args)
    );
  },

  submitAssignBid(args) {
    return api.post(`${invoke}`, this.buildPayload("submitAssignBid", args));
  },

  getAssignmentResults(args) {
    return api.post(
      `${query}`,
      this.buildPayload("getAssignmentResults", args)
    );
  }
};
