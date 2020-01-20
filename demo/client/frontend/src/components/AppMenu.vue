<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-navigation-drawer :clipped="true" :dark="true" permanent app>
    <v-list-item>
      <v-list-item-content>
        <v-list-item-title class="title">{{ userName }} </v-list-item-title>
        <v-list-item-subtitle>
          {{ userRole }}
        </v-list-item-subtitle>
      </v-list-item-content>
    </v-list-item>

    <v-divider></v-divider>

    <v-list dense nav>
      <v-list-item
        v-for="item in menuItems"
        :key="item.title"
        link
        :to="item.link"
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
    menuItems(state) {
      let bidderItems = [
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

      if (state.auth.role === "auction.bidder") {
        if (state.auction.state === "done") {
          bidderItems.shift();
        }

        return bidderItems.concat(defaultItems);
      } else {
        return defaultItems;
      }
    }
  }),

  methods: {
    logout: function() {
      this.$store.dispatch("auth/logout").then(() => {
        this.$router.push("/login");
      });
    }
  }
};
</script>
