/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Vue from "vue";
import VueRouter from "vue-router";
import store from "./store";
import Login from "./views/Login";

import AuctionInfo from "./views/AuctionInfo";
import PlaceBid from "./views/PlaceBid";

import Debug from "./views/Debug";

Vue.use(VueRouter);

const routes = [
  { path: "/auction_info", component: AuctionInfo },
  { path: "/place_bid", component: PlaceBid },
  { path: "/login", component: Login },
  { path: "/debug", component: Debug },
  { path: "/", component: AuctionInfo },
  { path: "*", redirect: "/" }
];

const router = new VueRouter({
  routes // short for `routes: routes`
});

router.beforeEach((to, from, next) => {
  if (to.path === "/debug") {
    // just skip login for debug view
    next();
  } else if (!store.getters["auth/isLoggedIn"] && to.path !== "/login") {
    next("/login");
  } else {
    next();
  }
});

export default router;
