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
                >Created by:
                <v-chip class="mx-1" color="orange" label text-color="white">
                  <v-avatar class="mr-2">
                    <v-img
                      v-bind:src="getAvatar($options.filters.dn(auction.owner))"
                    ></v-img>
                  </v-avatar>
                  {{ auction.owner | dn }}
                </v-chip>
              </span>
              <span class="mr-2" v-if="auction.state === 'clock'">
                <v-chip class="mx-1" color="blue" label text-color="white">
                  Clock Phase
                </v-chip>

                <v-chip class="mx-1" color="blue" label text-color="white">
                  Round {{ auction.clockRound }}
                </v-chip>

                <v-chip
                  class="mx-1"
                  color="green"
                  label
                  text-color="white"
                  v-if="isOpen"
                >
                  Round Active
                </v-chip>
                <v-chip
                  class="mx-1"
                  color="red"
                  label
                  text-color="white"
                  v-else
                >
                  Round Inactive
                </v-chip>
              </span>

              <span class="mr-2" v-if="auction.state === 'assign'">
                <v-chip class="mx-1" color="purple" label text-color="white">
                  Assignment Phase
                </v-chip>
              </span>

              <span class="mr-2" v-if="auction.state === 'done'">
                <v-chip class="mx-1" color="red" label text-color="white">
                  Auction Closed
                </v-chip>
              </span>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row v-if="showResults">
        <v-col cols="12">
          <v-expand-transition>
            <AuctionWinner />
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
import AuctionWinner from "./AuctionWinner";

export default {
  name: "AuctionDetails",
  components: { AuctionWinner },

  data: () => ({
    isLoading: true
  }),

  computed: {
    ...mapState({
      auction: state => state.auction,
      auctionState: state => state.auction.state,
      isAuctioneer: state => state.auth.role === "auction.auctioneer",
      isOpen: state => state.auction.roundActive
    }),

    ...mapGetters({
      getAvatar: "users/avatarByName",
      getColor: "users/colorByName"
    }),

    showResults() {
      return this.auction.state === "done" && this.auction.results !== "";
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

  watch: {
    // whenever question changes, this function will run
    // eslint-disable-next-line no-unused-vars
    auctionState: function(newState, oldState) {
      if (newState === "done") {
        this.fetchAssignmentResults(this.auction.id).catch(err =>
          console.log(err)
        );
      }
    }
  },

  methods: {
    ...mapActions({
      loadAuction: "auction/LOAD_AUCTION",
      fetchAssignmentResults: "auction/UPDATE_RESULTS"
    })
  },

  mounted() {
    // fetch the auction state .. try 1 (might not exist yet, though)
    this.loadAuction(1)
      .then(() => {
        if (this.auction.state === "done") {
          this.fetchAssignmentResults(this.auction.id);
        }
      })
      .catch(err => {
        if (err.rc != 5) {
          // 5 = invalid param ~= no such auction yet ..
          console.log(err);
        }
      })
      .finally(() => (this.isLoading = false));
  },

  filters: {
    dn: function(owner) {
      //"CN=Auctioneer1,OU=user+OU=org1"
      let p = owner.dn.match(/^CN=(\w+),.*$/);
      if (p === null) {
        return owner;
      }
      return p[1];
    }
  }
};
</script>
