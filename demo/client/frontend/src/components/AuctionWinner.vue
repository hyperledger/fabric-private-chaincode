<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-card dark>
    <v-card-title
      ><v-icon large class="mr-4" color="yellow">fas fa-trophy</v-icon> Auction
      Results</v-card-title
    >
    <v-card-text class="pa-0">
      <v-simple-table>
        <template v-slot:default>
          <thead>
            <tr>
              <th width="25%">Shows # of Licenses per Territory</th>
              <th
                v-for="(item, index) in getBidderResults"
                :key="index"
                width="25%"
              >
                <v-chip
                  class="my-2"
                  :color="getColor(item.name)"
                  label
                  text-color="white"
                >
                  <v-avatar size="28" class="mr-2">
                    <v-img v-bind:src="getAvatar(item.name)" /> </v-avatar
                  >{{ item.name }}
                </v-chip>
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(territory, index) in auction.territories" :key="index">
              <td>{{ territory.name }}</td>
              <td v-for="(item, val, index) in getBidderResults" :key="index">
                <span v-for="(t_obj, index) in item.territories" :key="index">
                  <span v-if="territory.name === t_obj.name" class="ma-5">
                    {{ t_obj.channels || 0 }}
                  </span>
                </span>
              </td>
            </tr>
            <tr style="background-color: #616161">
              <td>
                <span class="font-weight-bold">Total Price</span>
              </td>
              <td v-for="(item, val, index) in getBidderResults" :key="index">
                <span class="font-weight-bold">$ {{ item.totalCosts }}</span>
              </td>
            </tr>
          </tbody>
        </template>
      </v-simple-table>
    </v-card-text>
  </v-card>
</template>

<script>
import { mapGetters, mapState } from "vuex";

export default {
  name: "AuctionWinner",

  computed: {
    ...mapState({
      auction: state => state.auction
    }),

    ...mapGetters({
      getAvatar: "users/avatarByName",
      getColor: "users/colorByName"
    }),

    getBidderResults: function() {
      // helper lookup
      const bidderLookup = new Map(
        this.auction.bidders.map(b => [b.id, b.displayName])
      );

      let b = this.auction.results.map(r => {
        let entry = {};

        // lets get bidder name
        entry.name = bidderLookup.get(r.bidderId);

        // territories
        let assignments = new Map(
          r.assignment.map(t => [
            t.territoryId,
            {
              channels: t.channels.length,
              cost: t.channels.reduce((sum, ch) => sum + ch.price, 0)
            }
          ])
        );

        entry.territories = this.auction.territories.map(t => {
          let a = assignments.get(t.id);
          if (a === undefined) {
            a = { channels: 0, cost: 0 };
          }
          return {
            name: t.name,
            channels: a.channels,
            cost: a.cost
          };
        });

        // get total cost per bidder
        entry.totalCosts = entry.territories.reduce(
          (sum, t) => sum + t.cost,
          0
        );

        return entry;
      });

      return b;
    }
  }
};
</script>
