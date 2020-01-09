/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Vue from "vue";
import Vuex from "vuex";
import auction from "./modules/auction";
import bid from "./modules/bid";
import auth from "./modules/auth";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    auction,
    bid,
    auth
  },
  strict: process.env.NODE_ENV !== "production"
});
