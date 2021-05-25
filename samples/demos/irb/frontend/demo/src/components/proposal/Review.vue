<template>
  <v-card
      class="mx-auto"
      max-width="950"
      v-on:keyup.esc="onClose"
  >
    <v-card-title class="font-weight-regular blue white--text">
      <v-select
          class="mx-2 mt-0 pt-0"
          v-model="selectedReviewer"
          :items="reviewers"
          item-text="name"
          dark
          hide-details
      >
        <template v-slot:item="{ item }">
          <v-list-item-avatar color="grey darken-3">
            <v-img :src=getAvatar(item.name)></v-img>
          </v-list-item-avatar>

          <v-list-item-content>
            <v-list-item-subtitle>Reviewer</v-list-item-subtitle>
            <v-list-item-title class="headline font-weight-regular">{{ item.name }}</v-list-item-title>
          </v-list-item-content>
        </template>

        <template v-slot:selection="{ item }">
          <v-list-item-avatar color="grey darken-3">
            <v-img :src=getAvatar(item.name)></v-img>
          </v-list-item-avatar>

          <v-list-item-content>
            <v-list-item-subtitle>Reviewer</v-list-item-subtitle>
            <v-list-item-title class="headline font-weight-regular">{{ item.name }}</v-list-item-title>
          </v-list-item-content>
        </template>

      </v-select>

      <!--        <v-col>-->
      <!--      <v-spacer/>-->

      <!--      <v-btn-->
      <!--          icon-->
      <!--          class="ma-1"-->
      <!--          plain-->
      <!--          @click="onClose"-->
      <!--      >-->
      <!--        <v-icon color="white">mdi-close</v-icon>-->
      <!--      </v-btn>-->
      <!--        </v-col>-->

    </v-card-title>

    <v-card-text>

      <Proposal :proposal="proposalToApprove" :isReadOnly="true" class="mb-2"/>
      <v-btn
          v-if="!verified"
          color="purple"
          dark
          block
          @click="verify"
      >
        <v-icon dark class="mr-2">
          mdi-shield-search
        </v-icon>
        Verify
      </v-btn>
      <v-btn
          v-if="verified"
          color="green"
          dark
          block
          @click="verify"
      >
        <v-icon dark class="mr-2">
          mdi-shield-check
        </v-icon>
        Verified
      </v-btn>
    </v-card-text>

    <v-divider></v-divider>
    <v-card-actions>
      <v-checkbox
          class="ml-2"
          v-model="approved"
          :label="`I approve.`"
      ></v-checkbox>
      <v-spacer></v-spacer>
      <v-btn
          class="ma-1"
          color="error"
          plain
          :disabled="approved"
          @click="onReject"
      >Reject
      </v-btn>
      <v-btn
          class="ma-1"
          :disabled="!readyToApprove"
          @click="onApprove"
      >Approve
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>
import Proposal from '@/components/proposal/Proposal.vue';
import {mapGetters} from 'vuex';

const defaultReviewer = {name: 'Alice'};

export default {
  name: 'ProposalApprove',
  components: {
    Proposal
  },
  props: {
    reviewProposalWithId: String,
  },
  data: () => ({
    verified: false,
    approved: false,
    selectedReviewer: defaultReviewer.name,
  }),

  mounted() {
  },
  computed: {
    readyToApprove() {
      return this.verified && this.approved;
    },

    proposalToApprove() {
      return this.$store.getters['proposal/getProposalWithId'](this.reviewProposalWithId);
    },

    ...mapGetters({
      getAvatar: 'users/avatarByName',
      reviewers: 'users/userNames',
    }),
  },
  methods: {
    verify() {
      this.verified = !this.verified;
    },

    onReject() {
      this.$store.dispatch('proposal/reject', {
        proposalId: this.reviewProposalWithId,
        reviewer: this.selectedReviewer,
      });

      this.closeReview();
    },

    onApprove() {
      this.$store.dispatch('proposal/approve', {
        proposalId: this.reviewProposalWithId,
        reviewer: this.selectedReviewer,
      });

      this.closeReview();
    },

    onClose() {
      this.closeReview();
    },

    closeReview() {
      this.$emit('close-review');
    }
  },
  filters: {},
};

</script>

<style scoped>

</style>
