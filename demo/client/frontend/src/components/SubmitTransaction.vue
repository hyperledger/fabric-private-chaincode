<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <div class="text-center">
    <v-dialog v-model="dialog" persistent width="600">
      <v-card>
        <v-card-title>Submit Transaction</v-card-title>
        <v-card-text>
          <v-stepper v-model="progress" class="elevation-0" alt-labels>
            <v-stepper-header>
              <v-stepper-step
                :complete="progress > 1"
                step="1"
                color="pink"
                edit-icon="fas fa-cube"
                >Endorsement</v-stepper-step
              >
              <v-divider />

              <v-stepper-step :complete="progress > 2" step="2" color="purple"
                >Ordering</v-stepper-step
              >
              <v-divider />

              <v-stepper-step
                :complete="progress > 3"
                step="3"
                color="deep-purple"
                >Validate</v-stepper-step
              >
            </v-stepper-header>
            <v-stepper-items>
              <v-stepper-content step="1">
                <div class="text-center headline">
                  <v-progress-circular
                    indeterminate
                    color="pink"
                    class="mr-4"
                  />
                  Executing chaincode ...
                </div>
              </v-stepper-content>

              <v-stepper-content step="2">
                <div class="text-center headline">
                  <v-progress-circular
                    indeterminate
                    color="purple"
                    class="mr-4"
                  />
                  Sending transaction to ordering ...
                </div>
              </v-stepper-content>

              <v-stepper-content step="3">
                <div class="text-center headline">
                  <v-progress-circular
                    indeterminate
                    color="deep-purple"
                    class="mr-4"
                  />
                  Validate transaction ...
                </div>
              </v-stepper-content>

              <v-stepper-content step="4">
                <div class="text-center headline">
                  <v-icon large color="green darken-2" class="mr-4"
                    >fas fa-check</v-icon
                  >
                  Transaction complete
                </div>
              </v-stepper-content>
            </v-stepper-items>
          </v-stepper>
        </v-card-text>
      </v-card>
    </v-dialog>
  </div>
</template>

<script>
import { mapState, mapActions } from "vuex";

export default {
  data() {
    return {
      progress: 1,
      ticker: ""
    };
  },

  computed: {
    ...mapState({
      dialog: state => state.transaction.inProgress
    })
  },

  methods: {
    ...mapActions({
      fetchAuction: "transaction/finish"
    }),

    reset() {
      this.progress = 1;
      clearInterval(this.ticker);
    }
  },

  watch: {
    dialog(val) {
      if (!val) {
        this.reset();
        return;
      }

      let that = this;
      this.ticker = setInterval(() => {
        that.progress = that.progress + 1;
        if (that.progress > 4) {
          that.fetchAuction();
        }
        // 1200 default
      }, 1200);
    }
  }
};
</script>
