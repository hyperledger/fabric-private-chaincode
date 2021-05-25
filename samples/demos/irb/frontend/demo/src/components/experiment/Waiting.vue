<template>
  <v-card
      class="mx-auto"
      max-width="450">
    <v-card-title class="headline font-weight-regular purple white--text">
      <v-icon class="mr-2" large color="white">mdi-rocket-launch</v-icon>
      <span class="headline">Waiting for Approvals</span>
    </v-card-title>
    <v-card-text>
      <v-container style="height: 400px;">
        <v-row
            class="fill-height"
            align-content="center"
            justify="center"
        >
          <v-col
              class="subtitle-1 text-center"
              cols="12"
          >
            <v-item-group multiple>
              <v-item
                  v-for="review in proposalToWatch.reviews"
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
                  <v-avatar size="64" color="grey darken-3">
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
                  <v-avatar size="62" color="grey darken-3">
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
                  <v-avatar size="62" color="grey darken-3">
                    <v-img :src="getAvatar(review.name)"></v-img>
                  </v-avatar>
                </v-badge>

              </v-item>
            </v-item-group>
          </v-col>
          <v-col
              class="subtitle-1 text-center"
              cols="8"
          >
            <div v-if="!isApproved">
              Patience, my young Padawan!
            </div>
            <div v-else>
              Your experiment has been approved!
              Please continue.
            </div>
          </v-col>
        </v-row>
      </v-container>
    </v-card-text>
    <v-divider></v-divider>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
          :disabled="!isApproved"
          class="ma-1"
          @click="next"
      >Next
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>

import {mapGetters} from 'vuex';

export default {
  name: 'WaitingForApprovals',
  components: {},
  props: {
    watchProposalWithId: String,
  },
  data: () => ({
  }),
  computed: {
    proposalToWatch() {
      let proposal = this.$store.getters['proposal/getProposalWithId'](this.watchProposalWithId);
      if (proposal === undefined) {
        return {
          reviews: [
            {name: 'Alice',},
            {name: 'Bob',},
            {name: 'Charly',},
          ]
        }
      }
      return proposal
    },
    isApproved() {
      return this.$store.getters['proposal/isApproved'](this.watchProposalWithId);
    },
    ...mapGetters({
      getAvatar: 'users/avatarByName',
    })
  },
  methods: {
    next() {
      this.$emit('next')
    }
  },
  filters: {},
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
