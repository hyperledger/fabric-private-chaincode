const defaultUsers = [
	{
		name: 'Alice',
		avatar: 'https://avataaars.io/?avatarStyle=Transparent&topType=LongHairStraightStrand&accessoriesType=Prescription01&hairColor=BrownDark&facialHairType=Blank&clotheType=BlazerShirt&eyeType=EyeRoll&eyebrowType=UpDownNatural&mouthType=Concerned&skinColor=Brown',
	},
	{
		name: 'Bob',
		avatar: 'https://avataaars.io/?avatarStyle=Transparent&topType=ShortHairShortCurly&accessoriesType=Prescription02&hairColor=Black&facialHairType=Blank&clotheType=Hoodie&clotheColor=White&eyeType=Default&eyebrowType=DefaultNatural&mouthType=Default&skinColor=Light'
	},
	{
		name: 'Charly',
		avatar: 'https://avataaars.io/?avatarStyle=Transparent&topType=ShortHairShortWaved&accessoriesType=Prescription01&hairColor=Black&facialHairType=BeardLight&facialHairColor=Blonde&clotheType=BlazerSweater&eyeType=Cry&eyebrowType=AngryNatural&mouthType=ScreamOpen&skinColor=Light',
	}
];

const emptyUser = {
	name: '',
	avatar: ''
};

const state = {
	users: defaultUsers
};

const getters = {
	userByName: state => name => {
		return state.users.find(a => a.name === name) || emptyUser;
	},

	userNames(state) {
		return state.users.map(user => {
			return {name: user.name};
		});
	},

	avatarByName: (state, getters) => name => {
		return getters.userByName(name).avatar;
	},

	colorByName: (state, getters) => name => {
		return getters.userByName(name).color;
	}
};

const actions = {};

const mutations = {};

export default {
	namespaced: true,
	state,
	getters,
	actions,
	mutations
};