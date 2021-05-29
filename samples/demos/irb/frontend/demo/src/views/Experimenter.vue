<template>
  <v-stepper
      :value="currentProgressStep"
      class="elevation-0"
  >
    <v-stepper-header class="elevation-0">
      <v-stepper-step
          :complete="currentProgressStep > 1"
          step="1"
          editable
      >
        Create Proposal
      </v-stepper-step>
      <v-divider></v-divider>

      <v-stepper-step
          :complete="currentProgressStep > 2"
          step="2"
          editable
      >
        Wait for Approvals
      </v-stepper-step>
      <v-divider></v-divider>

      <v-stepper-step
          :complete="currentProgressStep > 3"
          step="3"
          editable
      >
        Run Experiment
      </v-stepper-step>

    </v-stepper-header>

    <v-stepper-items>
      <v-stepper-content step="1">
        <ProposalCreate
            class="mb-2"
            v-on:submit-proposal="onSubmitProposal"
        />
      </v-stepper-content>

      <v-stepper-content step="2">
        <WaitingForApprovals
            class="mb-2"
            :watchProposalWithId="currentProposalId"
            v-on:next="onNextStep"
        />
      </v-stepper-content>

      <v-stepper-content step="3">
        <RunExperiment
            class="mb-2"
            :watchProposalWithId="currentProposalId"
        />
      </v-stepper-content>

    </v-stepper-items>
  </v-stepper>
</template>

<script>
// @ is an alias to /src
import ProposalCreate from '@/components/proposal/Create.vue';
import WaitingForApprovals from '@/components/experiment/Waiting.vue';
import RunExperiment from '@/components/experiment/Run.vue';

import {mapGetters} from 'vuex';
import axios from 'axios';

export default {
  name: 'Experimenter',
  components: {
    ProposalCreate,
    WaitingForApprovals,
    RunExperiment,
  },
  data: () => ({
    currentProposalId: '1',
    currentStudyId: '1',
  }),
  computed: {
    ...mapGetters({
      currentProgressStep: 'experiment/currentProgressStep',
    }),
  },
  methods: {
    onSubmitProposal: function (studyId) {
      // TODO get study ID
      // TODO get proposal ID aka experiment ID

      this.currentStudyId = studyId

      console.log('propose experiment');
      const BASE_URL = process.env.VUE_APP_BASEURL_EXPERIMENTER
      axios.post(BASE_URL + '/api/new-experiment', {
        // TODO remove this hack
        studyId: '1',
        // TODO replace currentStudyId with currentProposalId
        experimentId: this.currentStudyId,
      }).then(response => {
        this.onNextStep();
        console.log(response);
      });
    },
    onNextStep: function () {
      this.$store.dispatch('experiment/nextStep');
    }
  },
};
</script>
