<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-navigation-drawer :clipped="true" :dark="true" permanent app>
    <v-list-item>
      <v-list-item-avatar>
        <v-img v-bind:src="userAvatar" />
      </v-list-item-avatar>
      <v-list-item-content>
        <v-list-item-title class="title">{{ userName }} </v-list-item-title>
        <v-list-item-subtitle>
          {{ userRole }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>

    <v-divider />

    <v-list dense nav>
      <v-list-item
        v-for="item in menuItems"
        :key="item.title"
        :to="item.link"
        link
      >
        <v-list-item-icon>
          <v-icon>{{ item.icon }}</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>

    <template v-slot:append>
      <div class="pa-2">
        <v-btn block @click="logout">Logout</v-btn>
      </div>
      <v-divider />
      <div>
        <v-card class="mx-auto" max-width="344">
          <v-img v-bind:src="'./img/fabric-logo.png'" class="mx-4 mt-4" />
          <v-card-title class="subtitle-1"
            >Fabric Private Chaincode</v-card-title
          >
          <v-card-subtitle class="caption">
            Enables the execution of chaincodes using Intel(R) SGX for
            Hyperledger Fabric.</v-card-subtitle
          >
        </v-card>
      </div>
    </template>
  </v-navigation-drawer>
</template>

<script>
import { mapState } from "vuex";

export default {
  name: "BidderMenu",

  computed: mapState({
    userName: state => state.auth.name,
    userRole: state => state.auth.role.replace("auction.", "").toUpperCase(),
    userAvatar: state => state.auth.avatar,
    isAuctionDone: state => state.auction.state.toString() === "done",
    isBidder: state => state.auth.role.toString() === "auction.bidder",

    menuItems() {
      let items = [
        { title: "Place Bids", link: "/place_bid", icon: "fa-hammer" },
        { title: "Bid Summary", link: "", icon: "fa-list-alt" },
        { title: "My Results", link: "", icon: "fa-chart-line" }
      ];

      let defaultItems = [
        {
          title: "Auction Info",
          link: "/auction_info",
          icon: "fa-info-circle"
        },
        { title: "Auction History", link: "", icon: "fa-calendar" },
        { title: "Settings", link: "", icon: "fa-cog" },
        { title: "Help", link: "", icon: "fa-question-circle" }
      ];

      if (!this.isBidder) {
        return defaultItems;
      }

      return items.concat(defaultItems);
    }
  }),

  methods: {
    logout() {
      this.$store
        .dispatch("auth/logout")
        .then(() => this.$router.push("/login"));
    }
  }
};
</script>
