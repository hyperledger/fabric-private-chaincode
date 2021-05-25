<template>
  <v-navigation-drawer app permanent expand-on-hover>

    <v-list
        nav
        dense
    >
      <v-list-item
          v-for="item in menuItems"
          :key="item.title"
          :to="item.link"
          link
      >
        <v-list-item-icon>
          <v-icon :color="item.color">{{ item.icon }}</v-icon>
        </v-list-item-icon>

        <v-list-item-content>
          <v-list-item-title>{{ item.title }}</v-list-item-title>
        </v-list-item-content>
      </v-list-item>
    </v-list>

  </v-navigation-drawer>
</template>

<script>
export default {

  computed: {
    menuItems() {
      let items = [];
      let defaultItems = [
        {title: 'About', link: '/', icon: 'mdi-home'},
        {title: 'Consenter', link: '/consenter', icon: 'mdi-database-arrow-up', color: 'green'},
        {title: 'Experimenter', link: '/experimenter', icon: 'mdi-rocket-launch', color: 'purple'},
        {title: 'Approver', link: '/approver', icon: 'mdi-account-multiple-check', color: 'blue'},
      ];
      return items.concat(defaultItems);
    },
  },
};
</script>