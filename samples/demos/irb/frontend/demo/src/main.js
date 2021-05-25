import Vue from 'vue';
import App from './App.vue';
import store from "./store";
import vuetify from './plugins/vuetify';
import router from './router';

Vue.config.productionTip = false

new Vue({
  vuetify,
  store,
  router,
  render: h => h(App)
}).$mount('#app')
