<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-toolbar color="primary" dark>
      <v-toolbar-title>Auction Actions</v-toolbar-title>
      <v-spacer></v-spacer>

      <v-toolbar-items>
        <CreateAuctionForm></CreateAuctionForm>
        <v-btn dark text @click="onClickEndRound">End round</v-btn>
        <v-btn dark text @click="onClickStartNextRound">Start next round</v-btn>
      </v-toolbar-items>
    </v-toolbar>
  </div>
</template>

<script>
import { mapState } from "vuex";
import CreateAuctionForm from "./CreateAuctionForm";

export default {
  name: "AuctionActions",
  components: { CreateAuctionForm },

  computed: {
    auction_id() {
      return { auctionId: this.auction.id };
    },

    ...mapState({
      auction: state => state.auction
    })
  },

  methods: {
    onClickEndRound() {
      this.$store.dispatch("auction/END_ROUND", this.auction_id);
    },

    onClickStartNextRound() {
      this.$store.dispatch("auction/START_NEXT_ROUND", this.auction_id);
    }
  }
};
</script>
