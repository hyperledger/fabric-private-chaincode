<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-row>
      <v-col cols="12">Clock Bidding</v-col>
    </v-row>

    <BiddingInfo
      :eligibility="eligibility"
      :requested-activity="requestedActivity"
      :required-activity="requiredActivity"
      :total-commitment="totalCommitment"
    >
    </BiddingInfo>

    <v-row>
      <v-col cols="12">
        <v-card>
          <v-card-title>
            <v-spacer></v-spacer>
            <v-text-field
              v-model="search"
              append-icon="fa-search"
              label="Search"
              single-line
              hide-details
            ></v-text-field>
          </v-card-title>

          <v-card-text class="pa-0">
            <v-data-table
              v-if="auction.currentRound === 1"
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
                    ></v-text-field>
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

              <template v-slot:item.clockPrice="props">
                $ {{ props.item.clockPrice }}
              </template>

              <template v-slot:item.price="props">
                <v-edit-dialog
                  :return-value.sync="props.item.price"
                  large
                  @save="save"
                >
                  <div>{{ props.item.price }}</div>
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
                    ></v-text-field>
                  </template>
                </v-edit-dialog>
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
                    ></v-text-field>
                  </template>
                </v-edit-dialog>
              </template>
            </v-data-table>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>

    <v-row>
      <v-col cols="12" right>
        <div class="d-flex justify-end">
          <v-btn color="primary" @click="prepareBid">Submit your bid</v-btn>
        </div>
      </v-col>
    </v-row>

    <v-overlay :value="waitingOverlay">
      <v-progress-circular indeterminate size="64"></v-progress-circular>
    </v-overlay>

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
          <v-spacer></v-spacer>
          <v-btn color="red darken-1" text @click="confirmDialog = false"
            >Cancel</v-btn
          >
          <v-btn color="green darken-1" text @click="submitBid">Submit</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { mapState } from "vuex";
import BiddingInfo from "./BiddingInfo";
import auction from "@/api/auction";
import axios from "axios";

export default {
  components: { BiddingInfo },
  data() {
    return {
      search: "",
      confirmDialog: false,
      waitingOverlay: false,
      territories: [],
      totalCommitment: 0,
      requestedActivity: 0,
      currentBid: null,
      defaultBid: {
        auctionId: "",
        round: "",
        bids: []
      }
    };
  },

  watch: {
    // this is a helper function to emulate some waiting :D
    waitingOverlay(val) {
      val &&
        setTimeout(() => {
          this.waitingOverlay = false;
        }, 1500);
    }
  },

  computed: {
    tableHeader: function() {
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
      const bidder = this.auction.bidders.find(
        bidder => bidder.displayName === this.auth.name
      );
      return this.auction.initialEligibilities.find(
        el => el.bidderId === bidder.id
      ).number;
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

      const request = {
        auctionId: this.auction.id || 1,
        round: this.auction.clockRound || 1
      };

      let that = this;
      axios
        .all([
          auction.getRoundInfo(request),
          auction.getBidderRoundResults(request)
        ])
        .then(
          axios.spread(function(roundInfo, roundResult) {
            if (roundInfo.data.status.rc !== 0) {
              console.log("error" + roundInfo.data.status.msg);
              return;
            }

            roundInfo.data.response.prices.forEach(p => {
              let i = that.territories.findIndex(t => t.id === p.terId);
              if (i > -1) {
                that.territories[i].minPrice = p.minPrice;
                that.territories[i].clockPrice = p.clockPrice;
              }
            });

            // do this only when we expect a result from getBidderRoundResults
            if (request.round > 1) {
              if (roundResult.data.status.rc !== 0) {
                console.log("error" + roundResult.data.status.msg);
                return;
              }

              roundResult.data.response.result.forEach(p => {
                let i = that.territories.findIndex(t => t.id === p.terId);
                if (i > -1) {
                  that.territories[i].price = p.postedPrice;
                  that.territories[i].demand = p.excessDemand;
                  that.territories[i].quantity = p.processedLicenses;
                }
              });

              // TODO set eligibility
            }
          })
        )
        .catch(err => console.log(err));
    },

    save() {
      this.totalCommitment = 0;
      this.requestedActivity = 0;

      this.territories.map(t => {
        this.totalCommitment += Number(t.quantity || 0) * Number(t.minPrice);
        this.requestedActivity += Number(t.quantity || 0);
      });
    },

    prepareBid: function() {
      this.currentBid = {
        auctionId: this.auction.id,
        round: this.auction.currentRound,
        bids: [],
        totalQuantity: 0,
        totalPrice: 0
      };

      this.territories
        .filter(t => t.quantity > 0)
        .map(t => {
          return {
            terId: t.id,
            territory: t.name,
            price: Number(t.price) || Number(t.minPrice),
            qty: Number(t.quantity)
          };
        })
        .forEach(l => {
          this.currentBid.bids.push(l);
          this.currentBid.totalQuantity += Number(l.qty);
          this.currentBid.totalPrice += Number(l.price);
        });

      this.confirmDialog = true;
    },

    submitBid: function() {
      this.confirmDialog = false;
      this.waitingOverlay = true;

      let bid = {
        auctionId: this.currentBid.auctionId,
        round: this.currentBid.round,
        bids: this.currentBid.bids.map(b => {
          return {
            terId: b.terId,
            price: b.price,
            qty: b.qty
          };
        })
      };

      this.$store.dispatch("bid/submitBid", bid).then(() => {
        this.currentBid = null;
      });
    }
  }
};
</script>
