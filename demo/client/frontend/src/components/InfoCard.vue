<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-card dark>
    <div class="d-flex flex-no-wrap justify-space-between">
      <div>
        <v-card-title>{{ title }}</v-card-title>
        <v-card-subtitle class="headline">{{ value }}</v-card-subtitle>
      </div>
    </div>
  </v-card>
</template>

<script>
export default {
  props: ["title", "value"]
};
</script>
