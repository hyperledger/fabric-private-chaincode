<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-navigation-drawer
    v-model="primaryDrawer.model"
    :clipped="true"
    :dark="true"
    width="400"
    app
  >
    <v-timeline dense class="mx-auto">
      <v-timeline-item
        color="purple lighten-2"
        fill-dot
        right
        icon="fa-cube"
        max-width="200"
        v-for="block in primaryDrawer.blocks.slice().reverse()"
        v-bind:key="block.id"
        v-bind:block="block"
      >
        <v-card max-width="260">
          <v-card-title class="purple lighten-2">
            <v-icon small left>fa-lock</v-icon>
            <span class="title font-weight-light" v-html="block.title">></span>
          </v-card-title>

          <v-container>
            <v-row>
              <v-col>
                some block content info
                <v-icon dark class="mr-4"> mdi-magnify </v-icon>
              </v-col>
            </v-row>
          </v-container>
        </v-card>
      </v-timeline-item>
    </v-timeline>
  </v-navigation-drawer>
</template>

<script>
export default {
  name: "Blockchain viewer",
  props: ["primaryDrawer"]
};
</script>
