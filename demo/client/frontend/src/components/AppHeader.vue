<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-app-bar :clipped-left="true" app>
    <v-toolbar-title class="headline text-uppercase">
      <strong>Fabric Private Chaincode</strong>
      <span class="font-weight-light"> - Auction Demo</span>
    </v-toolbar-title>

    <v-spacer />
    <!--    <v-toolbar-items>-->
    <!--      <v-menu offset-y v-if="auction.name">-->
    <!--        <template v-slot:activator="{ on }">-->
    <!--          <v-btn color="primary" dark v-on="on">-->
    <!--            {{ auctionState }}-->
    <!--          </v-btn>-->
    <!--        </template>-->
    <!--        <v-list>-->
    <!--          <v-list-item-->
    <!--            v-for="s in auctionStates"-->
    <!--            :key="s.key"-->
    <!--            @click="UPDATE_AUCTION_STATE(s.key)"-->
    <!--          >-->
    <!--            <v-list-item-title>{{ s.value }}</v-list-item-title>-->
    <!--          </v-list-item>-->
    <!--        </v-list>-->
    <!--      </v-menu>-->

    <!--      <v-menu offset-y v-if="auction.name && auction.state === 'clock'">-->
    <!--        <template v-slot:activator="{ on }">-->
    <!--          <v-btn color="secondary" dark v-on="on">-->
    <!--            Round {{ auction.clockRound }}-->
    <!--          </v-btn>-->
    <!--        </template>-->
    <!--        <v-list>-->
    <!--          <v-list-item-->
    <!--            v-for="index in 4"-->
    <!--            :key="index"-->
    <!--            @click="UPDATE_CLOCK_ROUND(index)"-->
    <!--          >-->
    <!--            <v-list-item-title>Round {{ index }}</v-list-item-title>-->
    <!--          </v-list-item>-->
    <!--        </v-list>-->
    <!--      </v-menu>-->
    <!--      &lt;!&ndash;      <v-btn text>{{ currentTime }}</v-btn>&ndash;&gt;-->
    <!--    </v-toolbar-items>-->
  </v-app-bar>
</template>

<script>
import { mapState, mapActions } from "vuex";
import moment from "moment";

export default {
  name: "AuctionHeader",
  data: function() {
    return {
      currentTime: null,
      auctionStates: [
        { key: "clock", value: "Clock Phase" },
        { key: "assign", value: "Assignment Phase" },
        { key: "done", value: "Done" },
        { key: "failed_fsr", value: "failed_FSR" }
      ]
    };
  },
  computed: {
    auctionState: function() {
      return this.auctionStates.find(s => s.key === this.auction.state).value;
    },
    ...mapState({
      auction: state => state.auction
    })
  },
  methods: {
    updateCurrentTime() {
      this.currentTime = moment().format("MM/DD/YYYY, h:mm:ss A z");
    },
    ...mapActions("auction", ["UPDATE_AUCTION_STATE", "UPDATE_CLOCK_ROUND"])
  },
  beforeDestroy() {
    clearInterval(this.$options.interval);
  },
  mounted() {
    this.updateCurrentTime();
    this.$options.interval = setInterval(this.updateCurrentTime, 1000);
  }
};
</script>
