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
      title: 'Presumptive diagnosis of Nephritis-the renal pelvis origin',
      studyId: '1',
      description: 'This experiment uses patient data to perform the presumptive diagnosis of Nephritis-the renal pelvis origin using Machine Learning Model.\n\nMore details on: https://archive.ics.uci.edu/ml/datasets/Acute+Inflammations',
      requestor: 'Sample Research Corp.',
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
          .then(() => this.$emit('submit-proposal', this.newProposal.studyId))
          .catch(err => console.error(err));
    }
  },
  filters: {},
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
