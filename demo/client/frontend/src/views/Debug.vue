<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container fluid>
    <v-app-bar app dark>
      <v-toolbar-title class="headline text-uppercase">
        <strong>Fabric Private Chaincode</strong>
        <span class="font-weight-light"> - Demo Dashboard</span>
      </v-toolbar-title>
    </v-app-bar>

    <v-row justify="center">
      <v-col cols="12" sm="12" md="10" lg="8" xl="6">
        <v-content>
          <v-container fluid grid-list-xl>
            <v-card>
              <v-tabs v-model="tab" background-color="blue-grey" dark>
                <v-tabs-slider color="light-blue lighten-1" />

                <v-tab href="#tab-1">
                  <v-icon class="mr-4">fas fa-cubes</v-icon>
                  Ledger
                </v-tab>

                <v-tab href="#tab-2">
                  <v-icon class="mr-4">fas fa-database</v-icon>
                  State
                </v-tab>

                <v-spacer />

                <v-tab href="#tab-3">
                  <v-icon class="mr-4">fas fa-gamepad</v-icon>
                </v-tab>
              </v-tabs>

              <v-tabs-items v-model="tab">
                <v-tab-item :value="'tab-1'">
                  <v-card flat>
                    <LedgerHistory />
                  </v-card>
                </v-tab-item>

                <v-tab-item :value="'tab-2'">
                  <v-card flat>
                    <LedgerTable />
                  </v-card>
                </v-tab-item>

                <v-tab-item :value="'tab-3'">
                  <v-card flat class="text-center">
                    <v-btn @click="resetDemo" class="ma-8">Reset demo</v-btn>
                  </v-card>
                </v-tab-item>
              </v-tabs-items>
            </v-card>
          </v-container>
        </v-content>
      </v-col>
    </v-row>
    <Footer />
  </v-container>
</template>

<script>
import LedgerTable from "../components/LedgerTable";
import LedgerHistory from "../components/LedgerHistory";
import Footer from "../components/Footer";
import Demo from "@/api/demo";

import { mapActions } from "vuex";

export default {
  name: "Debug",

  // Note that this view is currently only supported using the mock server
  // TODO implement ledger API for the gateway in order to make this view work

  components: {
    LedgerHistory,
    LedgerTable,
    Footer
  },

  data: () => ({
    tab: null
  }),

  created() {
    document.title = "Demo Dashboard";
  },

  methods: {
    ...mapActions({
      fetchState: "ledger/fetchState",
      newTransactionEvent: "ledger/newTransactionEvent",
      clearLedger: "ledger/clear"
    }),

    resetDemo() {
      // cleanup
      this.clearLedger();

      // go to first tab
      this.tab = "'tab-1'";

      //  wait a bit then restart
      setTimeout(() =>
        Demo.start().catch(err => console.log("error: " + err), 1000)
      );
    }
  }
};
</script>
