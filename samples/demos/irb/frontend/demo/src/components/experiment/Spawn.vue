<template>
  <v-card
      class="mx-auto"
      max-width="450">
    <v-card-title class="headline font-weight-regular purple white--text">
      <v-icon class="mr-2" large color="white">mdi-rocket-launch</v-icon>
      <span class="headline">Spawn experiment instance</span>
    </v-card-title>
    <v-card-text class="mt-6">
      <v-container>
        <v-row justify="center">
          <v-col
              class="text-center"
              cols="12"
          >
            <v-textarea
                v-model="proposalToWatch.codeIdentity"
                label="Code Identity"
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>

            <v-btn
                color="primary"
                x-large
                @click="startLaunch"
            >
              Launch Instance
            </v-btn>


          </v-col>
          <v-col cols="6"
                 v-if="isLaunching">
            <v-progress-linear

                color="deep-purple accent-4"
                indeterminate
                rounded
                height="6"
            ></v-progress-linear>
          </v-col>
          <v-col v-if="isLaunched" cols="12" class="text-center"
          >
            <v-textarea
                v-model="proposalToWatch.codeIdentity"
                label="Instance Public Key"
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>
            <v-textarea
                v-model="attestation"
                label="Attestation Evidence"
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>
            <div>
              Your experiment has been started!
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
          :disabled="!isLaunched"
          class="ma-1"
          @click="next"
      >Next
      </v-btn>
    </v-card-actions>
  </v-card>
</template>

<script>

export default {
  name: 'Spawn',
  components: {},
  props: {
    watchProposalWithId: String,
  },
  data: () => ({
    isLaunching: false,
    isLaunched: false,
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
        }
      }
      return proposal
    },
  },
  methods: {
    startLaunch() {
      this.isLaunching = true;

      setTimeout(() => {
        console.log('launched');
        this.isLaunched = true;
        this.isLaunching = false;
      }, 2500);
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
