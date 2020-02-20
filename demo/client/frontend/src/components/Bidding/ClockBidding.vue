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
      <v-col cols="12">Clock Bidding</v-col>
    </v-row>

    <div v-if="auction.id === ''">
      <v-alert prominent type="warning" class="mt-4">
        No auction running
      </v-alert>
    </div>

    <div v-else>
      <BiddingInfo
        :eligibility="eligibility"
        :requested-activity="requestedActivity"
        :required-activity="requiredActivity"
        :total-commitment="totalCommitment"
      />

      <v-row v-if="alertError || alertSuccess">
        <v-col cols="12">
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
            prominent
            dismissible
            transition="fade-transition"
            type="success"
            class="mt-4"
          >
            {{ successMsg }}
          </v-alert>
        </v-col>
      </v-row>

      <v-row>
        <v-col cols="12">
          <v-card>
            <v-card-title>
              <v-spacer />
              <v-text-field
                v-model="search"
                append-icon="fa-search"
                label="Search"
                single-line
                hide-details
              />
            </v-card-title>

            <v-card-text class="pa-0">
              <v-data-table
                v-if="auction.clockRound === 1"
                :headers="tableHeader"
                :items="territories"
                :search="search"
                hide-default-footer
              >
                <template v-slot:item.minPrice="props">
                  $ {{ props.item.minPrice }}
                </template>

                <template v-slot:item.supply="props">
                  {{ props.item.channels.length }}
                </template>

                <template v-slot:item.quantity="props">
                  <v-edit-dialog
                    :return-value.sync="props.item.quantity"
                    large
                    @save="save"
                  >
                    <div>{{ props.item.quantity || 0 }}</div>

                    <template v-slot:input>
                      <div class="mt-4 title">Update quantity</div>
                    </template>
                    <template v-slot:input>
                      <v-text-field
                        v-model="props.item.quantity"
                        type="number"
                        :rules="[
                          q =>
                            (q >= 0 && q <= props.item.channels.length) ||
                            'Invalid quantity'
                        ]"
                        label="Edit"
                        single-line
                        autofocus
                      />
                    </template>
                  </v-edit-dialog>
                </template>
              </v-data-table>

              <v-data-table
                v-else
                :headers="tableHeader"
                :items="territories"
                :search="search"
                hide-default-footer
              >
                <template v-slot:item.minPrice="props">
                  $ {{ props.item.minPrice }}
                </template>

                <template v-slot:item.supply="props">
                  {{ props.item.channels.length }}
                </template>

                <template v-slot:item.demand="props">
                  {{ props.item.demand || 0 }}
                </template>

                <template v-slot:item.clockPrice="props">
                  $ {{ props.item.clockPrice }}
                </template>

                <template v-slot:item.price="props">
                  <v-edit-dialog
                    :return-value.sync="props.value"
                    large
                    @save="save"
                  >
                    <div :key="newPrice">$ {{ props.item.price || 0 }}</div>
                    <template v-slot:input>
                      <div class="mt-4 title">Update bid price</div>
                    </template>
                    <template v-slot:input>
                      <v-text-field
                        v-model="props.item.price"
                        type="number"
                        :rules="[
                          q =>
                            (q >= props.item.minPrice &&
                              q <= props.item.clockPrice) ||
                            'Invalid price'
                        ]"
                        label="Edit"
                        single-line
                        autofocus
                      />
                    </template>
                  </v-edit-dialog>
                </template>

                <template v-slot:item.quantity="props">
                  <v-edit-dialog
                    :return-value.sync="props.value"
                    large
                    @save="save"
                  >
                    <div :key="newQuantity">{{ props.item.quantity || 0 }}</div>
                    <template v-slot:input>
                      <div class="mt-4 title">Update quantity</div>
                    </template>
                    <template v-slot:input>
                      <v-text-field
                        v-model="props.item.quantity"
                        type="number"
                        :rules="[
                          q =>
                            (q >= 0 && q <= props.item.channels.length) ||
                            'Invalid quantity'
                        ]"
                        label="Edit"
                        single-line
                        autofocus
                      />
                    </template>
                  </v-edit-dialog>
                </template>
              </v-data-table>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>

      <v-row justify="end">
        <v-col md="auto">
          <v-btn color="primary" @click="prepareBid">Submit your bid</v-btn>
        </v-col>
      </v-row>
    </div>

    <v-dialog v-if="currentBid" v-model="confirmDialog" max-width="400">
      <v-card>
        <v-card-title class="headline">Confirm your clock bid</v-card-title>
        <v-card-text>
          <v-simple-table>
            <template v-slot:default>
              <thead>
                <tr>
                  <th class="text-left">Territory</th>
                  <th class="text-left">Price</th>
                  <th class="text-left">Quantity</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="license in currentBid.bids" :key="license.territory">
                  <td>{{ license.territory }}</td>
                  <td>$ {{ license.price }}</td>
                  <td>{{ license.qty }}</td>
                </tr>
                <tr>
                  <th class="text-left">Total</th>
                  <th class="text-left">$ {{ currentBid.totalPrice }}</th>
                  <th class="text-left">{{ currentBid.totalQuantity }}</th>
                </tr>
              </tbody>
            </template>
          </v-simple-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="red darken-1" text @click="confirmDialog = false"
            >Cancel</v-btn
          >
          <v-btn color="green darken-1" text @click="submitBid">Submit</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <SubmitTransaction />
  </div>
</template>

<script>
import axios from "axios";
import { mapState } from "vuex";
import BiddingInfo from "./BiddingInfo";
import auction from "@/api/auction";
import SubmitTransaction from "../SubmitTransaction";
import helpers from "@/store/helpers";

export default {
  components: { BiddingInfo, SubmitTransaction },
  data() {
    return {
      isLoading: true,
      search: "",
      confirmDialog: false,
      // waitingOverlay: false,
      territories: [],
      totalCommitment: 0,
      requestedActivity: 0,
      newQuantity: 0,
      newPrice: 0,
      currentBid: null,
      defaultBid: {
        auctionId: "",
        round: "",
        bids: []
      },
      alertError: false,
      errMsg: "",
      successMsg: "",
      alertSuccess: false
    };
  },

  computed: {
    tableHeader() {
      if (this.auction.clockRound === 1) {
        return [
          { text: "Territory", value: "name" },
          { text: "Supply", value: "supply" },
          { text: "Opening Price", value: "minPrice" },
          { text: "Quantity", value: "quantity" }
        ];
      } else {
        return [
          { text: "Territory", value: "name" },
          { text: "Supply", value: "supply" },
          { text: "Agg. Demand", value: "demand" },
          { text: "Min Price", value: "minPrice" },
          { text: "Clock Price", value: "clockPrice" },
          { text: "Bid Price", value: "price" },
          { text: "Quantity", value: "quantity" }
        ];
      }
    },

    eligibility() {
      const bidder = this.bidder();
      if (bidder === undefined) {
        return 0;
      }

      const eligibility = this.auction.initialEligibilities.find(
        el => el.bidderId === bidder.id
      );

      return eligibility.number;
    },

    requiredActivity() {
      return (
        (this.eligibility * this.auction.activityRequirementPercentage) / 100
      );
    },

    ...mapState({
      auction: state => state.auction,
      auth: state => state.auth
    })
  },

  created() {
    this.initialize();
  },

  methods: {
    initialize() {
      this.territories = JSON.parse(JSON.stringify(this.auction.territories));

      const that = this;

      const roundInfoRequest = {
        auctionId: this.auction.id || 1,
        round: this.auction.clockRound || 1
      };

      // in the first round we only get the round info
      if (this.auction.clockRound === 1) {
        auction
          .getRoundInfo(roundInfoRequest)
          .then(resp => helpers.checkStatus(resp.data))
          .then(roundInfo =>
            roundInfo.prices.forEach(p => {
              let i = that.territories.findIndex(t => t.id === p.terId);
              if (i > -1) {
                that.territories[i].minPrice = p.minPrice;
                that.territories[i].clockPrice = p.clockPrice;
              }
            })
          )
          .catch(err => console.log(err))
          .finally(() => (this.isLoading = false));
        return;
      }

      const roundResultRequest = {
        auctionId: this.auction.id || 1,
        round: (this.auction.clockRound || 1) - 1
      };

      axios
        .all([
          auction
            .getRoundInfo(roundInfoRequest)
            .then(resp => helpers.checkStatus(resp.data)),
          auction
            .getBidderRoundResults(roundResultRequest)
            .then(resp => helpers.checkStatus(resp.data))
        ])
        .then(
          axios.spread(function(roundInfo, roundResult) {
            roundInfo.prices.forEach(p => {
              let i = that.territories.findIndex(t => t.id === p.terId);
              if (i > -1) {
                that.territories[i].minPrice = p.minPrice;
                that.territories[i].clockPrice = p.clockPrice;
              }
            });

            // do this only when we expect a result from getBidderRoundResults
            roundResult.result.forEach(p => {
              let i = that.territories.findIndex(t => t.id === p.terId);
              if (i > -1) {
                that.territories[i].price = p.postedPrice;
                that.territories[i].demand = p.excessDemand;
                that.territories[i].quantity = p.processedLicenses;
              }
            });

            // TODO set eligibility
          })
        )
        .catch(err => console.log(err))
        .finally(() => (this.isLoading = false));
    },

    bidder() {
      return this.auction.bidders.find(
        bidder => bidder.displayName === this.auth.name
      );
    },

    save() {
      this.totalCommitment = 0;
      this.requestedActivity = 0;

      this.territories.map(t => {
        this.totalCommitment += Number(t.quantity || 0) * Number(t.minPrice);
        this.requestedActivity += Number(t.quantity || 0);
        this.newQuantity = Number(t.quantity || 0);
        this.newPrice = Number(t.price);
      });
    },

    prepareBid() {
      this.currentBid = {
        auctionId: this.auction.id,
        round: this.auction.clockRound,
        bids: [],
        totalQuantity: 0,
        totalPrice: 0
      };

      this.territories
        .filter(t => t.quantity > 0)
        .map(t => ({
          terId: t.id,
          territory: t.name,
          price: Number(t.price) || Number(t.minPrice),
          qty: Number(t.quantity)
        }))
        .forEach(l => {
          this.currentBid.bids.push(l);
          this.currentBid.totalQuantity += Number(l.qty);
          this.currentBid.totalPrice += Number(l.price) * Number(l.qty);
        });

      this.confirmDialog = true;
    },

    resetCurrentBid() {
      this.currentBid = null;
    },

    submitBid() {
      this.confirmDialog = false;

      let bid = {
        auctionId: this.currentBid.auctionId,
        bidder: this.auth.name,
        round: this.currentBid.round,
        bids: this.currentBid.bids.map(b => ({
          terId: b.terId,
          price: b.price,
          qty: b.qty
        }))
      };

      this.$store
        .dispatch("transaction/submit")
        .then(() => this.$store.dispatch("bid/submitBid", bid))
        .then(() => this.resetCurrentBid())
        .then(() => this.onSuccess("Bid successfully submitted"))
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
