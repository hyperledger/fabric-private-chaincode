/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Vue from "vue";
import VueRouter from "vue-router";
import store from "./store";
import Login from "./views/Login";
import AuctionInfo from "./views/AuctionInfo";
import Results from "./views/MyResults";
import BidSummary from "./views/BidSummary";
import PlaceBid from "./views/PlaceBid";

Vue.use(VueRouter);

const routes = [
  { path: "/auction_info", component: AuctionInfo },
  { path: "/my_results", component: Results },
  { path: "/bid_summary", component: BidSummary },
  { path: "/place_bid", component: PlaceBid },
  { path: "/login", component: Login },
  { path: "/", component: AuctionInfo },
  { path: "*", redirect: "/" }
];

const router = new VueRouter({
  routes // short for `routes: routes`
});

router.beforeEach((to, from, next) => {
  if (!store.getters["auth/isLoggedIn"] && to.path !== "/login") {
    next("/login");
  } else {
    next();
  }
});

export default router;
