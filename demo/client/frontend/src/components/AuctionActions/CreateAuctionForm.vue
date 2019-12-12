<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-dialog v-model="dialog" persistent max-width="1024px">
    <template v-slot:activator="{ on }">
      <v-btn dark text v-on="on">Create Auction</v-btn>
    </template>
    <v-card>
      <v-card-title>
        <span class="headline">Create new Auction</span>
      </v-card-title>
      <v-card-text>
        <v-container>
          <v-row>
            <v-col cols="8">
              <v-text-field
                v-model="auction.name"
                label="Auction name*"
                required
              ></v-text-field>
            </v-col>

            <v-col cols="4">
              <v-text-field
                v-model="owner"
                label="Auction owner"
                disabled
              ></v-text-field>
            </v-col>

            <v-col cols="12">
              <TerritoryTable
                :territories="auction.territories"
                @update-territories="updateTerritories"
              >
              </TerritoryTable>
            </v-col>

            <v-col cols="12">
              <ParticipantsTable
                :participants="auction.bidders"
                :initialEligibilities="auction.initialEligibilities"
                @update-participants="updateBidders"
              >
              </ParticipantsTable>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-text-field
                v-model="auction.activityRequirementPercentage"
                label="Activity Requirement*"
                type="number"
                required
              ></v-text-field>
            </v-col>
            <v-col cols="12" sm="6" md="4">
              <v-text-field
                v-model="auction.clockPriceIncrementPercentage"
                label="Clock Price Increment*"
                type="number"
                required
              ></v-text-field>
            </v-col>
          </v-row>
        </v-container>
        <small>*indicates required field</small>
      </v-card-text>

      <v-card-actions>
        <v-spacer></v-spacer>
        <v-btn color="red darken-1" text @click="onClickCancel">Cancel</v-btn>
        <v-btn color="green darken-1" text @click="onClickSubmit">Submit</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script>
import { mapState } from "vuex";
import TerritoryTable from "../TerritoryTable";
import ParticipantsTable from "../ParticipantsTable";
import Auction from "@/api/auction";

export default {
  components: { ParticipantsTable, TerritoryTable },

  data: () => ({
    dialog: false,
    auction: {
      name: ""
    }
  }),

  created() {
    this.initialize();
  },

  watch: {
    dialog(val) {
      !val || this.onOpen();
    }
  },

  computed: mapState({
    owner: state => state.auth.name
  }),

  methods: {
    initialize() {
      Auction.getDefaultAuction()
        .then(response => {
          this.auction = response.data;
        })
        .catch(err => console.log(err));
    },

    onClickSubmit() {
      this.$store.dispatch("auction/NEW_AUCTION", this.auction);
      this.onClickCancel();
    },

    onClickCancel() {
      this.dialog = false;
    },

    onOpen() {
      this.initialize();
    },

    updateBidders(participants, eligibilities) {
      this.auction.bidders = JSON.parse(JSON.stringify(participants));
      this.auction.initialEligibilities = JSON.parse(
        JSON.stringify(eligibilities)
      );
    },

    updateTerritories(territories) {
      this.auction.territories = JSON.parse(JSON.stringify(territories));
    }
  }
};
</script>
