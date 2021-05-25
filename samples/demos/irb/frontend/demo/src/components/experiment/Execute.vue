<template>
  <v-card
      class="mx-auto"
      max-width="450">
    <v-card-title class="headline font-weight-regular purple white--text">
      <v-icon class="mr-2" large color="white">mdi-rocket-launch</v-icon>
      <span class="headline">Provision and Execute</span>
    </v-card-title>
    <v-card-text class="mt-6">
      <v-container>
        <v-row justify="center">
          <v-col
              class="text-center"
              cols="12"
          >
            <v-btn
                color="primary"
                x-large
                @click="startLaunch"
            >
              Execute
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
                v-model="result"
                label="Result"
                outlined
                :readonly="true"
                rows="2"
            ></v-textarea>
            <div>
              Thank you for conducting valuable research!
            </div>
          </v-col>
        </v-row>
      </v-container>
    </v-card-text>
  </v-card>
</template>

<script>

export default {
  name: 'Execute',
  components: {},
  props: {
    watchProposalWithId: String,
  },
  data: () => ({
    isLaunching: false,
    isLaunched: false,
    result: 'Not sure but my educated guess is: Cat'
  }),
  computed: {},
  methods: {
    startLaunch() {
      this.isLaunching = true;

      setTimeout(() => {
        console.log('launched');
        this.isLaunched = true;
        this.isLaunching = false;
      }, 3000);
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
