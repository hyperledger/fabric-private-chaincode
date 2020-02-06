/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

import Login from "@/api/login";

const state = {
  users: []
};

const emptyUser = {
  id: "",
  approle: "",
  color: "",
  avatar: ""
};

const userColors = ["light-blue", "deep-orange", "lime", "orange"];

const userAvatars = [
  "https://avataaars.io/?avatarStyle=Circle&topType=LongHairFrida&accessoriesType=Kurt&hairColor=Red&facialHairType=BeardLight&facialHairColor=BrownDark&clotheType=GraphicShirt&clotheColor=Gray01&graphicType=Skull&eyeType=Wink&eyebrowType=RaisedExcitedNatural&mouthType=Disbelief&skinColor=Brown",
  "https://avataaars.io/?avatarStyle=Circle&topType=ShortHairFrizzle&accessoriesType=Prescription02&hairColor=Black&facialHairType=MoustacheMagnum&facialHairColor=BrownDark&clotheType=BlazerSweater&clotheColor=Black&eyeType=Default&eyebrowType=FlatNatural&mouthType=Default&skinColor=Tanned",
  "https://avataaars.io/?avatarStyle=Circle&topType=LongHairMiaWallace&accessoriesType=Sunglasses&hairColor=BlondeGolden&facialHairType=Blank&clotheType=BlazerSweater&eyeType=Surprised&eyebrowType=RaisedExcited&mouthType=Smile&skinColor=Pale",
  "https://avataaars.io/?avatarStyle=Circle&topType=LongHairStraight&accessoriesType=Blank&hairColor=BrownDark&facialHairType=Blank&clotheType=BlazerShirt&eyeType=Default&eyebrowType=Default&mouthType=Default&skinColor=Light"
];

const getters = {
  userByName: state => name => {
    return state.users.find(a => a.id === name) || emptyUser;
  },

  userNames(state) {
    return state.users.map(user => user.id);
  },

  avatarByName: (state, getters) => name => {
    return getters.userByName(name).avatar;
  },

  colorByName: (state, getters) => name => {
    return getters.userByName(name).color;
  }
};

const actions = {
  fetchUsers({ commit }) {
    return Login.getRegisteredUsers()
      .then(response => response.data)
      .then(users => {
        let i = 0;
        return users.map(user => {
          return {
            id: user.id,
            approle: user.approle,
            avatar: userAvatars[i],
            color: userColors[i++]
          };
        });
      })
      .then(users => commit("setUsers", users));
  }
};

const mutations = {
  setUsers(state, users) {
    state.users = users;
  }
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
};
