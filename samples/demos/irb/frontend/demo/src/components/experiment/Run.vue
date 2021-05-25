<template>
  <v-card
      class="mx-auto"
      max-width="700">
    <v-card-title class="headline font-weight-regular purple white--text">
      <v-icon class="mr-2" large color="white">mdi-rocket-launch</v-icon>
      <span class="headline">Run Experiment</span>
    </v-card-title>
    <v-card-text class="mt-6">
      <v-container>

        <!-- Experiment details-->
        <v-row justify="center">
          <v-col class="text-center" cols="12">
            <v-text-field
                v-model="proposalToWatch.title"
                label="Title"
                outlined
                :readonly="true"
            ></v-text-field>
            <v-textarea
                v-model="proposalToWatch.codeIdentity"
                label="Code Identity"
                hide-details
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>
          </v-col>
        </v-row>

        <!-- Instance details-->
        <v-row justify="center">
          <v-col class="text-center" cols="6" v-if="!isLaunched">
            <v-btn
                color="primary"
                x-large
                @click="launchInstance"
            >
              <v-icon dark class="mr-2"> mdi-shield-lock</v-icon>
              Launch Instance
            </v-btn>
            <v-progress-linear
                v-if="isLaunching"
                color="deep-purple accent-4"
                class="mt-4"
                indeterminate
                rounded
                height="6"
            ></v-progress-linear>
          </v-col>

          <v-col v-else cols="12" class="text-center">
            <v-textarea
                v-model="publicKey"
                label="Public Key"
                outlined
                :readonly="true"
                rows="5"
            ></v-textarea>
            <v-textarea
                v-model="attestation"
                label="Attestation Evidence"
                hide-details
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>
          </v-col>
        </v-row>

        <!-- Experiment results -->
        <v-row justify="center" v-if="isLaunched">
          <v-col class="text-center" cols="6" v-if="!isDone">
            <v-btn
                color="primary"
                x-large
                @click="startExperiment"
            >
              <v-icon dark class="mr-2"> mdi-atom</v-icon>
              Start Experiment
            </v-btn>
            <v-progress-linear
                v-if="isExecuting"
                color="deep-purple accent-4"
                class="mt-4"
                indeterminate
                rounded
                height="6"
            ></v-progress-linear>
          </v-col>

          <v-col class="text-center" cols="12" v-if="isDone">
            <v-textarea
                v-model="result"
                label="Result"
                hide-details
                outlined
                :readonly="true"
                rows="4"
            ></v-textarea>
          </v-col>
        </v-row>

      </v-container>
    </v-card-text>
    <!--    <v-divider></v-divider>-->
    <!--    <v-card-actions>-->
    <!--      <v-spacer></v-spacer>-->
    <!--      <v-btn-->
    <!--          :disabled="!isLaunched"-->
    <!--          class="ma-1"-->
    <!--          @click="next"-->
    <!--      >Next-->
    <!--      </v-btn>-->
    <!--    </v-card-actions>-->
  </v-card>
</template>

<script>

import axios from 'axios';

export default {
  name: 'Spawn',
  components: {},
  props: {
    watchProposalWithId: String,
  },
  data: () => ({
    isLaunching: false,
    isLaunched: false,
    isExecuting: false,
    isDone: false,
    codeIdentity: '8ADE75BFDC8B76E01D56127A5CB01B6BC3006ADC48D3A41192ED8E949DB55DDB',
    publicKey: 'enclave public key',
    attestation: 'some attestation information'
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
        };
      }
      return proposal;
    },
  },
  methods: {
    launchInstance() {
      this.isLaunching = true;

      let that = this;
      axios.post('http://localhost:4000/api/launch', {}).then(response => {
        console.log('launched');
        that.isLaunched = true;
        that.isLaunching = false;
        console.log(response.data);
        that.publicKey = response.data.PublicKey;
        that.attestation = response.data.Attestation;
      });
    },
    startExperiment() {
      this.isExecuting = true;

      let that = this;
      axios.post('http://localhost:4000/api/execute', {}).then(response => {
        console.log('launched');
        that.isDone = true;
        that.isExecuting = false;
        console.log(response.data);
        that.result = response.data;
      });
    },
    next() {
      this.$emit('next');
    }
  },
  filters: {},
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
