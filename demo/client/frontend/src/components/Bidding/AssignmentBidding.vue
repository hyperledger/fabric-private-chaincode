<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container fluid>
    <v-row>
      <v-col cols="12">Assignment Bidding</v-col>
    </v-row>

    <v-row>
      <v-col cols="12">
        <v-card class="mx-auto">
          <v-tabs v-model="tab" background-color="orange darken-3" dark text>
            <v-tab
              v-for="territory in territories"
              :key="territory.name"
              :href="`#tab-${territory.name}`"
            >
              {{ territory.name }}
            </v-tab>

            <v-tab-item
              v-for="territory in territories"
              :key="territory.name"
              :value="'tab-' + territory.name"
            >
              <v-simple-table>
                <template v-slot:default>
                  <thead>
                    <tr>
                      <th class="text-left">Impairment Adjusted Price</th>
                      <th class="text-left">Channels</th>
                      <th class="text-left">Status</th>
                      <th class="text-left">Bid Price</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="options in territory.options" :key="options.id">
                      <td>$ {{ options.price }}</td>
                      <td>
                        <v-btn-toggle v-model="options.channels" multiple>
                          <v-btn
                            v-for="(item, index) in territory.channels"
                            v-bind:key="item.name"
                            color="primary"
                            disabled
                          >
                            {{ letters[index] }}
                          </v-btn>
                        </v-btn-toggle>
                      </td>
                      <td>{{ options.status }}</td>
                      <td>
                        <v-edit-dialog
                          :return-value.sync="options.bid"
                          large
                          @save="save"
                        >
                          <div>$ {{ options.bid }}</div>
                          <template v-slot:input>
                            <div class="mt-4 title">Place your bid</div>
                          </template>
                          <template v-slot:input>
                            <v-text-field
                              v-model="options.bid"
                              type="number"
                              :rules="[q => q >= 0 || 'Invalid price']"
                              label="Edit"
                              single-line
                              autofocus
                            />
                          </template>
                        </v-edit-dialog>
                      </td>
                    </tr>
                  </tbody>
                </template>
              </v-simple-table>
            </v-tab-item>
          </v-tabs>
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
      <v-progress-circular indeterminate size="64" />
    </v-overlay>

    <v-dialog v-model="confirmDialog" max-width="400">
      <v-card>
        <v-card-title class="headline"
          >Confirm your assignment bid</v-card-title
        >
        <v-card-text>
          <v-simple-table>
            <template v-slot:default>
              <thead>
                <tr>
                  <th class="text-left">Territory</th>
                  <th class="text-left">Channels</th>
                  <th class="text-left">Price</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="license in currentBid.licences"
                  :key="license.territory"
                >
                  <td>{{ license.territory }}</td>
                  <td>$ {{ license.price }}</td>
                  <td>{{ license.quantity }}</td>
                </tr>
              </tbody>
            </template>
          </v-simple-table>
        </v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn color="red darken-1" text @click="confirmDialog = false">
            Cancel
          </v-btn>
          <v-btn color="green darken-1" text @click="onClickSubmit">
            Submit
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </v-container>
</template>

<script>
// import axios from "axios";

export default {
  name: "Dashboard",
  data() {
    return {
      search: "",
      letters: "ABCDEFGHIJKLMNOPQRSTUVWXYZ".split(""),
      tab: null,
      confirmDialog: false,
      waitingOverlay: false,
      territories: null
    };
  },

  methods: {
    generateOptions: function(channelPrice, channels, wonChannels) {
      const numChannels = channels.length;
      let options = [];
      for (let i = 0; i <= numChannels - wonChannels; i++) {
        let chs = [];
        let price = 0;
        for (let j = i; j < i + wonChannels; j++) {
          price += (channelPrice * (100 - channels[j].impairment)) / 100;
          chs.push(j);
        }
        options.push({
          id: i,
          price: price,
          bid: 0,
          status: "-",
          channels: chs
        });
      }

      return options;
    },
    save() {},
    prepareBid: function() {
      this.currentBid = {
        id: "bla",
        licences: [],
        status: "submitted"
      };

      this.territories
        .filter(t => t.quantity > 0)
        .map(t => {
          return {
            territory: t.name,
            price: Number(t.openingPrice),
            quantity: Number(t.quantity)
          };
        })
        .map(l => this.currentBid.licenses.push(l));

      this.confirmDialog = true;
    },
    onClickSubmit: function() {
      this.confirmDialog = false;
      this.waitingOverlay = true;
      this.$store.dispatch("bid/submitBid", this.bid).then(() => {
        this.currentBid = null;
      });
    }
  },
  mounted() {
    // TODO: below URL does not exist, so comment all the relevant
    // code to suppress spurious errors in console.
    // should also use process.env.VUE_APP_API_BASE_URL instead of localhost
    // axios.get("http://localhost:3000/assignment").then(response => {
    //   this.territories = response.data.territories;
    //   this.territories.map(t => {
    //     t.options = this.generateOptions(t.clockPrice, t.channels, t.licenses);
    //   });
    //   console.log(this.territories);
    // });
  }
};
</script>
