<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-row>
      <v-col cols="12">My Bid Summary</v-col>
    </v-row>
    <v-card>
      <v-card-text class="pa-0">
        <v-data-table :headers="tableHeader" :items="bids" hide-default-footer>
          <template v-slot:item.openingPrice="props">
            $ {{ props.item.openingPrice }}
          </template>

          <template v-slot:item.action="{ item }">
            <v-icon small class="mr-2" @click="onClickMoreDetails(item)"
              >fa-search</v-icon
            >
          </template>
        </v-data-table>
      </v-card-text>
    </v-card>
  </div>
</template>

<script>
import { mapState } from "vuex";

export default {
  name: "BidSummaryTable",
  data() {
    return {
      tableHeader: [
        { text: "id", value: "id" },
        { text: "Round", value: "round" },
        { text: "Total Price", value: "totalPrice" },
        { text: "Total Quantity", value: "totalQuantity" },
        { text: "Status", value: "status" },
        { text: "Actions", value: "action", sortable: false }
      ]
    };
  },
  computed: mapState({
    bids: state => state.bid.submittedBids
  }),
  methods: {
    onClickMoreDetails(item) {
      console.log(item);
    }
  }
};
</script>
