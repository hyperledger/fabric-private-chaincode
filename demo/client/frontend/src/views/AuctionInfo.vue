<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container fluid>
    <AuctionHeader />
    <BidderMenu />
    <AuctionActions v-if="isAuctioneer" />
    <AuctionDetails />
  </v-container>
</template>

<script>
import { mapState } from "vuex";
import AuctionHeader from "../components/AppHeader";
import BidderMenu from "../components/AppMenu";
import AuctionDetails from "../components/AuctionDetails";
import AuctionActions from "../components/AuctionActions/AuctionActions";

export default {
  name: "AuctionInfo",

  components: {
    AuctionHeader,
    BidderMenu,
    AuctionDetails, // this is our current default view
    AuctionActions
  },

  computed: {
    ...mapState({
      isAuctionOpen: state => !!state.auction.id,
      isAuctioneer: state => state.auth.role.toString() === "auction.auctioneer"
    })
  }
};
</script>
