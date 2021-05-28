<template>
  <v-container>

    <v-row class="mx-0 mt-0">
      <v-col cols="6" class="px-0">
        <v-text-field
            v-model="proposal.title"
            label="Title"
            outlined
            :readonly="isReadOnly"
        ></v-text-field>
      </v-col>
      <v-col cols="3" class="pl-4 pr-8">
        <v-text-field
            v-model="proposal.studyId"
            label="Study ID"
            outlined
            hide-details
            :readonly="isReadOnly"
        ></v-text-field>
      </v-col>

      <v-col cols="3" class="px-0">
        <v-text-field
            v-model="proposal.requestor"
            label="Requestor"
            outlined
            hide-details
            :readonly="isReadOnly"
        ></v-text-field>
      </v-col>
    </v-row>

    <v-row class="mx-0 mt-0">
      <v-col cols="9" class="px-0 pr-8">
        <v-textarea
            v-model="proposal.description"
            label="Description"
            auto-grow
            outlined
            hide-details
            :readonly="isReadOnly"
            rows="5"
        ></v-textarea>
      </v-col>

      <v-col cols="3" class="px-0 mx-0 pl-4">
        <v-radio-group
            class="mt-0"
            label="Organization Type"
            v-model="proposal.checkedAllowedUse"
            :readonly="isReadOnly"
            hide-details
        >
          <v-radio
              v-for="n in allowedUseItems"
              :key="n"
              :label="`${n}`"
              :value="n"
          ></v-radio>
        </v-radio-group>
      </v-col>
    </v-row>


    <v-row>
      <v-col cols=12>
        <span class="text-h6">Experiment details</span>
        <v-divider></v-divider>
        <CodeView :files="proposal.files"  v-on:update-files="onUpdateCodeIdentity" :isReadOnly="isReadOnly"></CodeView>

        <v-combobox
            v-model="proposal.selectedDataDomains"
            :items="dataDomainItems"
            multiple
            label="Input Data Domain"
            :readonly="isReadOnly"
        >
          <template v-slot:selection="data">
            <v-chip
                :key="JSON.stringify(data.item)"
                v-bind="data.attrs"
                :input-value="data.selected"
                :disabled="data.disabled"
                @click:close="data.parent.selectItem(data.item)"
            >
              <v-avatar
                  class="accent white--text"
                  left
                  v-text="data.item.slice(0, 1).toUpperCase()"
              ></v-avatar>
              {{ data.item }}
            </v-chip>
          </template>
        </v-combobox>

      </v-col>
    </v-row>
  </v-container>
</template>

<script>
import CodeView from '@/components/CodeView.vue';

import {mapState} from 'vuex';

export default {
  name: 'Proposal',
  components: {
    CodeView,
  },
  props: {
    isReadOnly: Boolean,
    proposal: Object,
  },
  data: () => ({
  }),
  computed: mapState({
    allowedUseItems: state => state.proposal.defaults.allowedUseItems,
    dataDomainItems: state => state.proposal.defaults.dataDomainItems,
  }),
  methods: {
    onUpdateCodeIdentity(newCodeIdentity) {
      this.proposal.codeIdentity = newCodeIdentity
    }
  },
  filters: {},
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

</style>
