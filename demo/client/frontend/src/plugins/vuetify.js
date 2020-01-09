/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Vue from "vue";
import Vuetify from "vuetify/lib";
import "@fortawesome/fontawesome-free/css/all.css"; // Ensure you are using css-loader

Vue.use(Vuetify);

const opts = {
  icons: {
    iconfont: "fa"
  }
};

export default new Vuetify(opts);
