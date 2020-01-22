<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div v-if="isLoading">
    <v-progress-linear
      :active="isLoading"
      indeterminate
      absolute
      top
      color="light-blue accent-4"
    />
  </div>

  <div v-else>
    <v-row>
      <v-col cols="auto">Auction Info</v-col>
    </v-row>

    <div v-if="auction.id === ''">
      <v-alert prominent type="warning" class="mt-4">
        No auction running
      </v-alert>
    </div>
    <div v-else>
      <v-row>
        <v-col cols="12">
          <v-card dark>
            <v-card-title class="title">{{ auction.name }}</v-card-title>
            <v-card-text class="px-4">
              <span class="mr-8"
                >Created by
                <v-chip class="mx-1" color="orange" label text-color="white">
                  <v-avatar class="mr-2">
                    <v-img v-bind:src="getAvatar(auction.owner)"></v-img>
                  </v-avatar>
                  {{ auction.owner }}
                </v-chip>
              </span>

              <span class="mr-8"
                >Status
                <v-chip
                  class="mx-1"
                  color="green"
                  label
                  text-color="white"
                  v-if="isOpen"
                >
                  Open
                </v-chip>
                <v-chip
                  class="mx-1"
                  color="red"
                  label
                  text-color="white"
                  v-else
                >
                  Closed
                </v-chip>
              </span>

              <span>
                Round
                <v-chip class="mx-1" color="green" label text-color="white">
                  {{ auction.clockRound }}
                </v-chip>
              </span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row v-if="showWinner">
        <v-col cols="12">
          <v-expand-transition>
            <v-card v-show="showWinner" dark>
              <div class="d-flex flex-no-wrap justify-space-between">
                <v-list-item two-line>
                  <v-list-item-avatar size="260" class="mb-8">
                    <img alt="Avatar" :src="getAvatar(auction.winner.bidder)" />
                  </v-list-item-avatar>
                  <v-list-item-content>
                    <div class="headline mb-8">The winner is ...</div>
                    <v-list-item-title class="display-4 mb-2">{{
                      auction.winner.bidder
                    }}</v-list-item-title>
                    <v-list-item-subtitle class="title"
                      >with a bid of $
                      {{ auction.winner.value }}</v-list-item-subtitle
                    >
                  </v-list-item-content>
                </v-list-item>
              </div>
            </v-card>
          </v-expand-transition>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="12" v-if="isAuctioneer">
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
                      <td>
                        <v-chip
                          class="my-2"
                          :color="getColor(item.displayName)"
                          label
                          text-color="white"
                        >
                          <v-avatar size="28" class="mr-2">
                            <v-img v-bind:src="getAvatar(item.displayName)" />
                          </v-avatar>
                          {{ item.displayName }}
                        </v-chip>
                      </td>
                      <td>{{ item.eligibility }}</td>
                      <td>{{ item.id }}</td>
                    </tr>
                  </tbody>
                </template>
              </v-simple-table>
            </v-card-text>
          </v-card>
        </v-col>

        <v-col cols="12">
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
        </v-col>
      </v-row>
    </div>
    <!--        <v-flex lg12>-->
    <!--          <v-card>-->
    <!--            <v-card-title>Debug</v-card-title>-->
    <!--            <v-card-text>-->
    <!--              <pre>{{ auction }}</pre>-->
    <!--            </v-card-text>-->
    <!--          </v-card>-->
    <!--        </v-flex>-->
  </div>
</template>

<script>
import { mapState, mapActions, mapGetters } from "vuex";

export default {
  name: "AuctionDetails",

  data: () => ({
    isLoading: true
  }),

  computed: {
    ...mapState({
      auction: state => state.auction,
      isAuctioneer: state => state.auth.role === "auction.auctioneer",
      isOpen: state => state.auction.roundActive
    }),

    ...mapGetters({
      getAvatar: "users/avatarByName",
      getColor: "users/colorByName"
    }),

    showWinner() {
      return this.isAuctioneer && this.auction.winner;
    },

    // todo this move to auction state as a getter
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

  methods: {
    ...mapActions({
      fetchAuction: "auction/LOAD_AUCTION"
    })
  },

  mounted() {
    // fetch the auction state .. try 1
    this.fetchAuction(1)
      .catch(err => console.log(err))
      .finally(() => (this.isLoading = false));
  }
};
</script>
