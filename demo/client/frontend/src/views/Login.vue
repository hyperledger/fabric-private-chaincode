<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-container fluid class="fill-height">
    <v-row align="center" justify="center">
      <v-col cols="12" sm="8" md="4">
        <v-card class="elevation-12">
          <v-toolbar color="primary" dark flat>
            <v-toolbar-title>Spectrum Auction Login</v-toolbar-title>
          </v-toolbar>
          <v-card-text>
            <v-form>
              <v-select
                v-model="selection"
                :items="userNames"
                filled
                label="Login"
                prepend-icon="fa-user"
              />

              <v-text-field
                id="password"
                label="Password"
                name="password"
                prepend-icon="fa-lock"
                type="password"
              />
            </v-form>
          </v-card-text>
          <v-card-actions>
            <v-spacer />
            <v-btn
              color="primary"
              @click="loginUser"
              :disabled="!somethingSelected"
              >Login</v-btn
            >
          </v-card-actions>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import { mapGetters } from "vuex";

export default {
  name: "Login",

  data: () => ({
    selection: ""
  }),

  mounted() {
    this.$store.dispatch("users/fetchUsers").catch(err => console.log(err));
  },

  computed: {
    ...mapGetters({
      userNames: "users/userNames",
      userByName: "users/userByName"
    }),

    somethingSelected() {
      return this.selection !== "";
    }
  },

  methods: {
    loginUser: function() {
      const user = this.userByName(this.selection);
      this.$store
        .dispatch("auth/login", user)
        .then(() => this.$router.push("/"))
        .catch(err => console.log(err));
    }
  }
};
</script>
