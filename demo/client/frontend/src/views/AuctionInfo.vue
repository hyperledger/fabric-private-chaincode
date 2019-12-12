<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <AuctionHeader></AuctionHeader>
    <BidderMenu></BidderMenu>
    <AuctionActions v-if="userRole === 'auction.auctioneer'"></AuctionActions>
    <AuctionDetails></AuctionDetails>
  </div>
</template>

<script>
import AuctionHeader from "../components/AppHeader";
import BidderMenu from "../components/AppMenu";
import AuctionDetails from "../components/AuctionDetails";
import AuctionActions from "../components/AuctionActions/AuctionActions";

import { mapState } from "vuex";

export default {
  name: "AuctionInfo",
  components: {
    AuctionHeader,
    BidderMenu,
    AuctionDetails, // this is our current default view
    AuctionActions
  },

  computed: mapState({
    userRole: state => state.auth.role
  })
};
</script>
