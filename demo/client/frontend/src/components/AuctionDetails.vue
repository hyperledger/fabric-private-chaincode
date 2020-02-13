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
            <v-card v-show="showResults" dark>
              <div class="d-flex flex-no-wrap justify-space-between">
                <v-list-item two-line>
                  <v-list-item-content>
                    <div class="headline mb-8">Auction Results</div>
                    <v-simple-table>
                      <template v-slot:default>
                        <thead>
                          <tr>
                            <th>Territory Name</th>
                            <th></th>
                            <th
                              v-for="(item, index) in getBidderResults"
                              :key="index"
                            >
                              {{ item.name }}
                            </th>
                          </tr>
                        </thead>
                        <tbody>
                          <tr
                            v-for="(territory, index) in auction.territories"
                            :key="index"
                          >
                            <td>{{ territory.name }}</td>
                            <td>
                              <tr>
                                <td># of Channels</td>
                              </tr>
                              <tr>
                                <td>Total cost</td>
                              </tr>
                            </td>

                            <td
                              v-for="(item, val, index) in getBidderResults"
                              :key="index"
                            >
                              <span
                                v-for="(t_obj, index) in item.territories"
                                :key="index"
                              >
                                <span
                                  v-if="territory.name === t_obj.name"
                                  class="justify-content-center"
                                >
                                  <tr>
                                    <td style="vertical-align:middle">
                                      {{ t_obj.channels }}
                                    </td>
                                  </tr>
                                  <tr>
                                    <td style="vertical-align:middle">
                                      ${{ t_obj.cost }}
                                    </td>
                                  </tr>
                                </span>
                              </span>
                            </td>
                          </tr>
                        </tbody>
                      </template>
                    </v-simple-table>
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
      return this.isAuctioneer && this.auction.state === "done";
    },

    getBidderResults: function() {
      var displayInfo = [];
      for (var i in this.auction.results) {
        var bidder_dic = {};
        for (var bidder in this.auction.bidders) {
          if (
            this.auction.bidders[bidder].id === this.auction.results[i].bidderId
          ) {
            bidder_dic.name = this.auction.bidders[bidder].displayName;
            break;
          }
        }
        var territoryList = [];
        for (var j in this.auction.results[i].assignment) {
          var territoryInfo = {};
          var tmp = "";
          for (var t in this.auction.territories) {
            if (
              this.auction.territories[t].id ===
              this.auction.results[i].assignment[j].territoryId
            ) {
              tmp = this.auction.territories[t].name;
              break;
            }
          }

          // getting channel prices
          var totalcost = 0;
          var c = 0;
          for (var k in this.auction.results[i].assignment[j].channels) {
            c = c + 1;
            totalcost =
              totalcost +
              this.auction.results[i].assignment[j].channels[k].price;
            console.log("Count " + c);
          }
          console.log("Total cost " + totalcost);
          var channelcount = this.auction.results[i].assignment[j].channels
            .length; // this is the count of channels
          territoryInfo.name = tmp;
          territoryInfo.channels = channelcount;
          territoryInfo.cost = totalcost;
          territoryList.push(territoryInfo);
        }
        bidder_dic.territories = territoryList; // setting territory names

        displayInfo.push(bidder_dic); // appending all the info for displaying in an array
      }

      return displayInfo;
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
        this.fetchAssignmetresults(1).catch(err => console.log(err));
      }
    }
  },

  methods: {
    ...mapActions({
      fetchAuction: "auction/LOAD_AUCTION",
      fetchAssignmetresults: "auction/UPDATE_RESULTS"
    })
  },

  mounted() {
    // fetch the auction state .. try 1
    this.fetchAuction(1)
      .then(this.fetchAssignmetresults(1))
      .catch(err => console.log(err))
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
