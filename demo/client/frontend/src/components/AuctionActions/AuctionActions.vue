<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-toolbar color="primary" dark>
      <v-toolbar-title>Auction Actions</v-toolbar-title>
      <v-spacer />

      <v-toolbar-items>
        <SubmitTransaction />
        <CreateAuctionForm v-on:error="onError" v-on:success="onSuccess" />
        <v-btn dark text @click="onClickStartNextRound">Start Round</v-btn>
        <v-btn dark text @click="onClickEndRound">End Round</v-btn>
      </v-toolbar-items>
    </v-toolbar>

    <v-alert
      v-model="alertError"
      prominent
      dismissible
      transition="fade-transition"
      type="error"
      class="mt-4"
    >
      {{ errMsg }}
    </v-alert>

    <v-alert
      v-model="alertSuccess"
      dismissible
      transition="fade-transition"
      type="success"
      class="mt-4"
    >
      {{ successMsg }}
    </v-alert>
  </div>
</template>

<script>
import { mapState } from "vuex";
import CreateAuctionForm from "./CreateAuctionForm";
import SubmitTransaction from "../SubmitTransaction";

export default {
  name: "AuctionActions",
  components: { CreateAuctionForm, SubmitTransaction },

  data() {
    return {
      alertError: false,
      errMsg: "",
      successMsg: "",
      alertSuccess: false
    };
  },

  computed: {
    ...mapState({
      auction: state => state.auction,
      auctionId: state => state.auction.id
    })
  },

  methods: {
    onClickEndRound() {
      this.$store
        .dispatch("transaction/submit")
        .then(() => this.$store.dispatch("auction/END_ROUND", this.auctionId))
        .then(() => this.onSuccess("Auction successfully closed"))
        .catch(err => this.onError(err));
    },

    onClickStartNextRound() {
      this.$store
        .dispatch("transaction/submit")
        .then(() => this.$store.dispatch("auction/NEXT_ROUND", this.auctionId))
        .then(() =>
          this.$store.dispatch("auction/LOAD_AUCTION", this.auctionId)
        )
        .catch(err => this.onError(err));
    },

    onSuccess(msg) {
      this.successMsg = msg;
      this.alertSuccess = true;
    },

    onError(error) {
      this.errMsg = error.message;
      this.alertError = true;
    }
  }
};
</script>
