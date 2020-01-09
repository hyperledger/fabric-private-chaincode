<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div>
    <v-row>
      <v-col cols="12">My Results</v-col>
    </v-row>

    <v-data-table
      :headers="headers"
      :items="bids"
      :items-per-page="5"
      class="elevation-1"
    ></v-data-table>

    <v-simple-table>
      <template v-slot:default>
        <thead>
          <tr>
            <th class="text-left">Geographic Parcel</th>
            <th class="text-left">Supply</th>
            <th class="text-left">Agg Demand</th>
            <th class="text-left">My Processed Demand</th>
            <th class="text-left">Round 1 Posted Price</th>
            <th class="text-left">Round 2 Clock Price</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in bids" :key="item.name">
            <td>{{ item.name }}</td>
            <td>{{ item.supply }}</td>
            <td>{{ item.demand }}</td>
            <td>{{ item.mybids }}</td>
            <td>${{ item.price }}</td>
            <td>${{ item.price }}</td>
          </tr>
        </tbody>
      </template>
    </v-simple-table>
  </div>
</template>

<script>
export default {
  name: "Dashboard",
  props: ["dashboard"],
  data() {
    return {
      bids: [
        {
          name: "Elbograd",
          supply: 10,
          demand: 4,
          mybids: 3,
          price: 10000
        },
        {
          name: "Mudberg",
          supply: 10,
          demand: 4,
          mybids: 3,
          price: 6000
        },
        {
          name: "Deserton",
          supply: 9,
          demand: 4,
          mybids: 3,
          price: 6000
        },
        {
          name: "Phlimsk",
          supply: 6,
          demand: 4,
          mybids: 3,
          price: 4000
        }
      ]
    };
  }
};
</script>
