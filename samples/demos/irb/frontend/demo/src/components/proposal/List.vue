<template>
  <v-card
      class="mx-auto"
      max-width="850">
    <v-card-title class="headline font-weight-regular blue white--text">
      <v-icon class="mr-2" large color="white">mdi-account-multiple-check</v-icon>
      <span class="headline">Experiment Proposals</span>
    </v-card-title>
    <v-card-text>
      <v-data-table
          :headers="headers"
          :items="proposals"
          :items-per-page="5"
      >
        <template v-slot:item.approvals="{ item }">
          <v-item-group multiple>
            <v-item
                v-for="review in item.reviews"
                :key="review.name"
                class="mx-1"
            >
              <v-badge
                  v-if="review.status === 'approved'"
                  avatar
                  bordered
                  overlap
                  icon="mdi-check"
              >
                <v-avatar size="32" color="grey darken-3">
                  <v-img :src="getAvatar(review.name)"></v-img>
                </v-avatar>
              </v-badge>

              <v-badge
                  v-else-if="review.status === 'rejected'"
                  avatar
                  bordered
                  overlap
                  color="error"
                  icon="mdi-close"
              >
                <v-avatar size="32" color="grey darken-3">
                  <v-img :src="getAvatar(review.name)"></v-img>
                </v-avatar>
              </v-badge>

              <v-badge
                  v-else
                  avatar
                  bordered
                  overlap
                  color="grey"
                  icon="mdi-help"
              >
                <v-avatar size="32" color="grey darken-3">
                  <v-img :src="getAvatar(review.name)"></v-img>
                </v-avatar>
              </v-badge>

            </v-item>
          </v-item-group>
        </template>

        <template v-slot:item.actions="{ item }">
          <v-btn
              color="blue-grey"
              class="ma-2 white--text"
              v-on:click="$emit('select-review', item.id)"
              text
          >
            <v-icon dark>
              mdi-file-find
            </v-icon>
            Review
          </v-btn>
        </template>
      </v-data-table>
    </v-card-text>
  </v-card>
</template>

<script>
import {mapGetters} from 'vuex';

export default {
  name: 'ProposalList',
  components: {},
  data: () => ({
    headers: [
      {text: 'Title', align: 'start', value: 'title',},
      {text: 'Requestor', value: 'requestor',},
      {text: 'Approvals', value: 'approvals', width: '160px'},
      {text: 'Status', value: 'status'},
      {text: '', value: 'actions', sortable: false},
      // {text: '', align: 'end', value: 'actions', sortable: false},
    ]
  }),
  computed: {
    ...mapGetters({
      proposals: 'proposal/getAllProposals',
      getAvatar: 'users/avatarByName',
    }),
  },
  methods: {},
  filters: {},
};
</script>