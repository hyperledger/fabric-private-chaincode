<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-card>
    <v-card-text>
      <div class="text-center ma-4" v-if="!hasTransactions">
        No transactions yet
      </div>
      <v-timeline v-else dense>
        <v-slide-x-reverse-transition group hide-on-leave>
          <v-timeline-item
            v-for="item in ledger"
            :key="item.uuid"
            :icon="txIcon"
            fill-dot
          >
            <v-card dark>
              <v-card-title>Transaction</v-card-title>
              <v-card-subtitle>{{ item.uuid }}</v-card-subtitle>
              <v-card-text>
                <v-chip class="ma-2" color="pink" label text-color="white">
                  <v-icon left small>fas fa-handshake</v-icon>
                  {{ item.chaincode }}
                </v-chip>

                <v-chip
                  v-if="item.creator"
                  class="ma-2"
                  :color="getColor(item.creator)"
                  label
                  text-color="white"
                >
                  <v-avatar class="mr-2">
                    <v-img v-bind:src="getAvatar(item.creator)"></v-img>
                  </v-avatar>
                  {{ item.creator }}
                </v-chip>

                <v-chip class="ma-2" color="primary" label text-color="white">
                  <v-icon left small>fas fa-terminal</v-icon>
                  {{ item.func }}
                </v-chip>

                <v-badge
                  icon="fas fa-lock"
                  color="deep-purple"
                  bottom
                  overlap
                  bordered
                  offset-y="18"
                  offset-x="18"
                  :value="isLocked(item.func)"
                >
                  <v-chip
                    class="ma-2"
                    color="deep-purple"
                    label
                    text-color="white"
                  >
                    <v-icon left small>fas fa-database</v-icon>
                    Data
                  </v-chip>
                </v-badge>

                <v-tooltip bottom color="grey">
                  <template v-slot:activator="{ on }">
                    <v-badge
                      icon="fas fa-certificate"
                      color="teal"
                      bottom
                      overlap
                      bordered
                      offset-y="18"
                      offset-x="18"
                    >
                      <v-chip
                        class="ma-2"
                        color="teal"
                        label
                        text-color="white"
                        v-on="on"
                      >
                        <v-icon left small>fas fa-scroll</v-icon>
                        Enclave Attestation
                      </v-chip>
                    </v-badge>
                  </template>
                  <div>
                    Attestation report: MRENCLAVE + Enclave Pk; Intel Signature
                  </div>
                </v-tooltip>

                <!--                <v-chip-->
                <!--                  class="ma-2"-->
                <!--                  color="green"-->
                <!--                  label-->
                <!--                  text-color="white"-->
                <!--                  v-if="item.is_valid"-->
                <!--                >-->
                <!--                  <v-icon left small>fas fa-check</v-icon>-->
                <!--                  Valid Tx-->
                <!--                </v-chip>-->

                <!--                <v-chip-->
                <!--                  class="ma-2"-->
                <!--                  color="red"-->
                <!--                  label-->
                <!--                  text-color="white"-->
                <!--                  v-else-->
                <!--                >-->
                <!--                  <v-icon left small>fas fa-times</v-icon>-->
                <!--                  Invalid Tx-->
                <!--                </v-chip>-->
              </v-card-text>
            </v-card>
          </v-timeline-item>
        </v-slide-x-reverse-transition>
      </v-timeline>
    </v-card-text>
  </v-card>
</template>

<script>
import { mapState, mapGetters, mapActions } from "vuex";

export default {
  data: () => ({
    dialog: false,
    isLoading: true,
    txIcon: "fas fa-cube"
  }),

  computed: {
    ...mapState({
      ledger: state => state.ledger.transactions
    }),

    ...mapGetters({
      hasTransactions: "ledger/hasTransactions",
      getAvatar: "users/avatarByName",
      getColor: "users/colorByName"
    })
  },

  methods: {
    isLocked(f) {
      const publicData = ["publishAssignmentResults"];
      return !publicData.includes(f);
    },

    ...mapActions({
      fetchUsers: "users/fetchUsers",
      fetchTransactions: "ledger/fetchTransactions"
    })
  },

  mounted() {
    // get avatars
    this.fetchUsers().catch(err => console.log(err));

    // first get transactions
    this.fetchTransactions().catch(err => console.log(err));
  }
};
</script>
