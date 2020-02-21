<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container fluid>
    <AuctionHeader />
    <BidderMenu />
    <ClockBidding v-if="isClockPhase" />
    <AssignmentBidding v-else-if="isAssignPhase" />
    <div v-else>
      <v-alert prominent type="warning" class="mt-4">
        No auction running
      </v-alert>
    </div>
  </v-container>
</template>

<script>
import { mapState } from "vuex";
import AuctionHeader from "../components/AppHeader";
import BidderMenu from "../components/AppMenu";
import ClockBidding from "../components/Bidding/ClockBidding";
import AssignmentBidding from "../components/Bidding/AssignmentBidding";

export default {
  name: "PlaceBid",

  components: {
    AuctionHeader,
    BidderMenu,
    ClockBidding,
    AssignmentBidding
  },

  computed: {
    ...mapState({
      isClockPhase: state =>
        state.auction.state !== undefined &&
        state.auction.state.toString() === "clock",
      isAssignPhase: state =>
        state.auction.state !== undefined &&
        state.auction.state.toString() === "assign"
    })
  }
};
</script>
