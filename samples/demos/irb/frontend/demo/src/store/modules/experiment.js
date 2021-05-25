

const emptyExperiment = {
	progressStep: 1,
	proposalId: '',
};

const state = {
	experiment: emptyExperiment
};

const getters = {
	currentProgressStep: state => {
		return state.experiment.progressStep
	},
};

const actions = {
	nextStep({commit}) {
		commit("nextStep");
	},
};

const mutations = {
	nextStep: state => {
		state.experiment.progressStep++
	}
};

export default {
	namespaced: true,
	state,
	getters,
	actions,
	mutations
};