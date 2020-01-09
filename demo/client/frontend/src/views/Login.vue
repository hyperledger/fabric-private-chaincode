<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-app id="inspire">
    <v-content>
      <v-container class="fill-height" fluid>
        <v-row align="center" justify="center">
          <v-col cols="12" sm="8" md="4">
            <v-card class="elevation-12">
              <v-toolbar color="primary" dark flat>
                <v-toolbar-title>Spectrum Auction Login</v-toolbar-title>
                <v-spacer></v-spacer>
              </v-toolbar>
              <v-card-text>
                <v-form>
                  <v-col class="d-flex" cols="12">
                    <v-select
                      v-model="selection"
                      :items="userNames"
                      filled
                      label="Registered Users"
                    ></v-select>
                  </v-col>
                  <v-text-field
                    id="password"
                    label="Password"
                    name="password"
                    prepend-icon="fa-lock"
                    type="password"
                  ></v-text-field>
                </v-form>
              </v-card-text>
              <v-card-actions>
                <v-spacer></v-spacer>
                <router-link v-if="selection != ''" to="/auction_info">
                  <v-btn color="primary" @click="loginUser">Login</v-btn>
                </router-link>
                <v-btn v-else color="primary" disabled>Login</v-btn>
              </v-card-actions>
            </v-card>
          </v-col>
        </v-row>
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
import Login from "@/api/login";

export default {
  name: "Login",

  data: () => ({
    userData: [],
    userNames: [],
    selection: ""
  }),

  mounted() {
    Login.getRegisteredUsers()
      .then(response => {
        this.userData = response.data;
        this.userNames = this.userData.map(el => el.id);
      })
      .catch(err => console.log(err));
  },

  methods: {
    loginUser: function() {
      let user = this.userData.find(({ id }) => id === this.selection);
      this.$store
        .dispatch("auth/login", user)
        .then(() => this.$router.push("/"));
    }
  }
};
</script>
