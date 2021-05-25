<template>
  <v-card
      class="mx-auto"
      max-width="950">
    <v-card-title class="headline font-weight-regular purple white--text">
      <v-icon class="mr-2" large color="white">mdi-rocket-launch</v-icon>
      <span class="headline">New Experiment</span>
    </v-card-title>
    <v-card-text>
      <Proposal :proposal="newProposal"/>
    </v-card-text>
    <v-divider></v-divider>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
          :disabled="uploadDisabled"
          class="ma-1"
          @click="submit"
      >Submit Proposal
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>
import Proposal from '@/components/proposal/Proposal.vue';

export default {
  name: 'ProposalCreate',
  components: {
    Proposal
  },
  data: () => ({
    newProposal: {
      title: 'The Woodman experiment',
      description: 'The Woodman set to work at once, and so sharp was his axe that the tree was soon chopped to the end. The Woodman set to work at once, and so sharp was his axe that the tree was soon chopped.',
      requestor: 'Fancy Research Corp.',
      checkedAllowedUse: 'Private',
      selectedDataDomains: ['Health'],
      files: [],
      codeIdentity: '',
    },
  }),
  computed: {
    uploadDisabled() {
      return false;
    }
  },
  methods: {
    submit() {
      this.$store
          .dispatch('proposal/submit', this.newProposal)
          .then(() => this.$emit('submit-proposal'))
          .catch(err => console.error(err));
    }
  },
  filters: {},
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
