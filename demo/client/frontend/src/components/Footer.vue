<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-footer dark app padless>
    <v-card class="flex" flat tile>
      <v-card-text class="py-2 white--text text-center">
        <v-icon small>fas fa-code</v-icon>
        <span class="mx-2">Source code and examples available on Github</span>
        <span class="mr-2">-</span>
        <strong
          ><a
            >https://github.com/hyperledger-labs/fabric-private-chaincode</a
          ></strong
        >
      </v-card-text>
    </v-card>
  </v-footer>
</template>

<script>
export default {
  name: "Footer"
};
</script>
