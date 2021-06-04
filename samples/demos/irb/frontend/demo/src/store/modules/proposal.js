const defaultsProposal = {
	title: 'The Woodman experiment',
	studyId: '1',
	description: 'The Woodman set to work at once, and so sharp was his axe that the tree was soon chopped to the end. The Woodman set to work at once, and so sharp was his axe that the tree was soon chopped.',
	requestor: 'Fancy Research Corp.',
	checkedAllowedUse: 'Private',
	selectedDataDomains: ['Health'],
	files: [],
	codeIdentity: '',
	reviews: [
		{name: 'Alice', status: 'approved'},
		{name: 'Bob', status: 'approved'},
		{name: 'Charly', status: 'approved'},
	],
	status: 'Proposed',
};

const emptyProposal = {
	id: "",
	title: '',
	studyId: '',
	description: '',
	requestor: '',
	checkedAllowedUse: '',
	selectedDataDomains: [],
	files: [],
	codeIdentity: '',
	reviews: [
		{name: 'Alice', status: 'pending'},
		{name: 'Bob', status: 'pending'},
		{name: 'Charly', status: 'pending'},
	],
	status: 'Proposed'
};

const state = {
	proposals: [],
	defaults: {
		allowedUseItems: [
			'Private',
			'Public',
			'Government',
		],
		dataDomainItems: [
			'Health',
			'Financial',
		],
		statusItems: [
			'Proposed',
			'Reviewed',
		],
	}
};

const getters = {
	getDefaultsProposal: () => {
		return defaultsProposal;
	},

	getAllProposals: state => {
		return state.proposals;
	},

	getProposalWithId: (state) => (id) => {
		return state.proposals.find(proposal => proposal.id === id)
	},

	isApproved: () => (id) => {
		let proposal = state.proposals.find(proposal => proposal.id === id)
		if (proposal === undefined) {
			return false;
		}
		return proposal.reviews.some(review => review.status === "approved")
	}
};

const actions = {
	submit({commit}, proposal) {
		commit("newProposal", proposal);
	},

	approve({commit}, {proposalId, reviewer}) {
		let proposal = state.proposals.find(proposal => proposal.id === proposalId)
		if (proposal === undefined) {
			console.error("cannot find proposal with id=" + proposalId)
			return // error
		}
		// make a copy here
		proposal = JSON.parse(JSON.stringify(proposal))

		let reviewPos = proposal.reviews.findIndex(review => review.name === reviewer )
		if (reviewPos === -1) {
			console.error("cannot find reviewer=" + reviewer)
			return // error
		}

		proposal.reviews[reviewPos].status = "approved"

		// check if we have enough approvals
		if (proposal.reviews.some(review => review.status === "approved")) {
			proposal.status = "Reviewed"
		}

		commit("updateProposal", proposal);
	},

	reject({commit}, {proposalId, reviewer}) {
		let proposal = state.proposals.find(proposal => proposal.id === proposalId)
		if (proposal === undefined) {
			console.error("cannot find proposal with id=" + proposalId)
			return // error
		}
		// make a copy here
		proposal = JSON.parse(JSON.stringify(proposal))

		let reviewPos = proposal.reviews.findIndex(review => review.name === reviewer )
		if (reviewPos === -1) {
			console.error("cannot find reviewer=" + reviewer)
			return // error
		}

		proposal.reviews[reviewPos].status = "rejected"
		commit("updateProposal", proposal);
	}
};

const mutations = {
	newProposal(state, proposal) {
		proposal = {
			...emptyProposal,
			...JSON.parse(JSON.stringify(proposal)),
		}
		proposal.id = (state.proposals.length + 1).toString()
		state.proposals.push(proposal);
	},

	updateProposal(state, updatedProposal) {
		let pos = state.proposals.findIndex(proposal => proposal.id === updatedProposal.id)
		if (pos === -1) {
			return // abort
		}

		// let's apply changes
		updatedProposal = {
			...JSON.parse(JSON.stringify(state.proposals[pos])),
			...JSON.parse(JSON.stringify(updatedProposal)),
		}

		// set updated
		state.proposals[pos] = updatedProposal
	},
};

export default {
	namespaced: true,
	state,
	getters,
	actions,
	mutations
};