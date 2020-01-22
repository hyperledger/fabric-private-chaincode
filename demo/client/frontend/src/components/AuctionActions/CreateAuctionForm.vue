<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-dialog v-model="dialog" persistent max-width="1024px">
    <template v-slot:activator="{ on }">
      <v-btn dark text v-on="on">Create</v-btn>
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
                v-model="newAuction.name"
                label="Auction name*"
                required
              />
            </v-col>

            <v-col cols="4">
              <v-text-field v-model="owner" label="Auction owner" disabled />
            </v-col>

            <v-col cols="12">
              <TerritoryTable
                :territories="newAuction.territories"
                @update-territories="updateTerritories"
              >
              </TerritoryTable>
            </v-col>

            <v-col cols="12">
              <ParticipantsTable
                :participants="newAuction.bidders"
                :initialEligibilities="newAuction.initialEligibilities"
                @update-participants="updateBidders"
              >
              </ParticipantsTable>
            </v-col>

            <v-col cols="12" sm="6" md="4">
              <v-text-field
                v-model="newAuction.activityRequirementPercentage"
                label="Activity Requirement*"
                type="number"
                required
              />
            </v-col>
            <v-col cols="12" sm="6" md="4">
              <v-text-field
                v-model="newAuction.clockPriceIncrementPercentage"
                label="Clock Price Increment*"
                type="number"
                required
              />
            </v-col>
          </v-row>
        </v-container>
        <small>*indicates required field</small>
      </v-card-text>

      <v-card-actions>
        <v-spacer />
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
    newAuction: {
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
    owner: state => state.auth.name,
    auction: state => state.auction
  }),

  methods: {
    initialize() {
      Auction.getDefaultAuction()
        .then(response => (this.newAuction = response.data))
        .catch(err => console.log(err));
    },

    onClickSubmit() {
      this.$store
        .dispatch("transaction/submit")
        .then(() =>
          this.$store.dispatch("auction/NEW_AUCTION", this.newAuction)
        )
        .then(() =>
          this.$store.dispatch("auction/LOAD_AUCTION", this.auction.id)
        )
        .then(() => this.$emit("success", "Auction successfully created"))
        .catch(error => this.$emit("error", error))
        .finally(() => this.onClickCancel());
    },

    onClickCancel() {
      this.dialog = false;
    },

    onOpen() {
      this.initialize();
    },

    updateBidders(participants, eligibilities) {
      this.newAuction.bidders = JSON.parse(JSON.stringify(participants));
      this.newAuction.initialEligibilities = JSON.parse(
        JSON.stringify(eligibilities)
      );
    },

    updateTerritories(territories) {
      this.newAuction.territories = JSON.parse(JSON.stringify(territories));
    }
  }
};
</script>
