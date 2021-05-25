import Vue from "vue";
import Vuex from "vuex";
import proposal from "./modules/proposal";
import experiment from "./modules/experiment";
import users from "./modules/users";

Vue.use(Vuex);

export default new Vuex.Store({
	modules: {
		proposal,
		experiment,
		users,
	},
	strict: process.env.NODE_ENV !== "production"
});