<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <template>
      <AuctionHeader></AuctionHeader>
      <BidderMenu></BidderMenu>
      <ClockBidding v-if="auction.state === 'clock'"
        >currentPhase=={{ auction.state }}</ClockBidding
      >
      <AssignmentBidding v-else></AssignmentBidding>
    </template>
  </div>
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
  computed: mapState({
    auction: state => state.auction
  })
};
</script>
