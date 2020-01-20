/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Vue from "vue";
import Vuex from "vuex";
import auction from "./modules/auction";
import bid from "./modules/bid";
import auth from "./modules/auth";
import transaction from "./modules/transaction";
import ledger from "./modules/ledger";
import users from "./modules/users";

Vue.use(Vuex);

export default new Vuex.Store({
  modules: {
    auction,
    bid,
    auth,
    transaction,
    ledger,
    users
  },
  strict: process.env.NODE_ENV !== "production"
});
