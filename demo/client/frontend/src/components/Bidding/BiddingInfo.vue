<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container grid-list-xl fluid px-0>
    <v-layout row wrap>
      <v-flex lg3 sm6 xs12>
        <InfoCard title="Total commitment" :value="'$ ' + totalCommitment" />
      </v-flex>

      <v-flex lg3 sm6 xs12>
        <InfoCard title="Eligibility" :value="eligibility" />
      </v-flex>

      <v-flex lg3 sm6 xs12>
        <InfoCard title="Required activity" :value="requiredActivity" />
      </v-flex>

      <v-flex lg3 sm6 xs12>
        <InfoCard title="Requested activity" :value="requestedActivity" />
      </v-flex>
    </v-layout>
  </v-container>
</template>

<script>
import InfoCard from "../InfoCard";

export default {
  components: { InfoCard },
  props: [
    "totalCommitment",
    "eligibility",
    "requiredActivity",
    "requestedActivity"
  ]
};
</script>
