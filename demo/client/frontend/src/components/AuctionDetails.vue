<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-row>
      <v-col cols="12">Auction Info</v-col>
    </v-row>

    <v-container grid-list-xl fluid px-0>
      <v-layout row wrap>
        <v-flex lg8>
          <v-card>
            <v-card-text class="pa-0">
              <div class="layout row ma-0">
                <div class="sm12 xs6 flex py-3">
                  <div class="headline">{{ auction.name }}</div>
                  <span class="caption">Created by {{ auction.owner }}</span>
                </div>
              </div>
            </v-card-text>
          </v-card>
        </v-flex>

        <v-flex lg6 sm12>
          <v-card>
            <v-card-title>Auction participants</v-card-title>
            <v-card-text class="pa-0">
              <v-simple-table>
                <template v-slot:default>
                  <thead>
                    <tr>
                      <th class="text-left">Name</th>
                      <th class="text-left">Eligibility</th>
                      <th class="text-left">Id</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in bidders" :key="item.name">
                      <td>{{ item.displayName }}</td>
                      <td>{{ item.eligibility }}</td>
                      <td>{{ item.id }}</td>
                    </tr>
                  </tbody>
                </template>
              </v-simple-table>
            </v-card-text>
          </v-card>
        </v-flex>

        <v-flex lg6 sm12>
          <v-card>
            <v-card-title>Territories</v-card-title>
            <v-card-text class="pa-0">
              <v-simple-table>
                <template v-slot:default>
                  <thead>
                    <tr>
                      <th class="text-left">Name</th>
                      <th class="text-left">Supply</th>
                      <th class="text-left">Opening bidding price</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="item in auction.territories" :key="item.name">
                      <td>{{ item.name }}</td>
                      <td>{{ item.channels.length }}</td>
                      <td>$ {{ item.minPrice }}</td>
                    </tr>
                  </tbody>
                </template>
              </v-simple-table>
            </v-card-text>
          </v-card>
        </v-flex>

        <v-flex lg12>
          <v-card>
            <v-card-title>Debug</v-card-title>
            <v-card-text>
              <pre>{{ auction }}</pre>
            </v-card-text>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
import { mapState } from "vuex";

export default {
  name: "AuctionDetails",

  computed: {
    ...mapState({
      auction: state => state.auction
    }),

    bidders() {
      return this.auction.bidders.map(bidder => {
        const container = {};
        Object.assign(container, bidder);
        Object.assign(
          container,
          this.auction.initialEligibilities
            .filter(y => y.bidderId === bidder.id)
            .map(el => ({ eligibility: el.number }))
            .shift()
        );
        return container;
      });
    }
  },

  mounted() {
    this.$store.dispatch("auction/LOAD_AUCTION", { auctionId: 1 });
  }
};
</script>
