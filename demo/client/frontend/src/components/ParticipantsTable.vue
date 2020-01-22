<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-data-table
    :headers="headers"
    :items="mergedData"
    class="elevation-1"
    hide-default-footer
  >
    <template v-slot:top>
      <v-toolbar flat color="white">
        <v-toolbar-title>Participants</v-toolbar-title>
        <v-spacer></v-spacer>
      </v-toolbar>
    </template>

    <template v-slot:item.channels="props">
      <span
        class="mr-2"
        v-for="channel in props.item.channels"
        v-bind:key="channel.id"
      >
        {{ channel.name }} = {{ channel.impairment }}%
      </span>
    </template>

    <template v-slot:item.action="{ item }">
      <v-icon small class="mr-2" @click="onClickEditItem(item)">
        fa-edit
      </v-icon>
      <v-icon small @click="onClickDeleteItem(item)"> fa-trash </v-icon>
    </template>

    <template v-slot:body.append="{ headers }">
      <tr>
        <td :colspan="headers.length" class="text-right">
          <v-dialog v-model="dialog" max-width="500px">
            <template v-slot:activator="{ on }">
              <v-btn icon v-on="on"><v-icon>fa-plus-circle</v-icon></v-btn>
            </template>
            <v-form ref="form" v-model="validInput">
              <v-card>
                <v-card-title
                  ><span class="headline">{{ formTitle }}</span></v-card-title
                >
                <v-card-text>
                  <v-container>
                    <v-row>
                      <v-col cols="12">
                        <v-text-field
                          v-model="editedItem.displayName"
                          label="Name"
                          :rules="nameRules"
                          required
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12">
                        <v-text-field
                          v-model="editedItem.principal.mspId"
                          label="MSP Id"
                          :rules="mspRules"
                          required
                        ></v-text-field>
                      </v-col>
                      <v-col cols="12">
                        <v-text-field
                          v-model="editedItem.principal.dn"
                          label="Dn"
                          :rules="dnRules"
                          required
                        ></v-text-field>
                      </v-col>
                      <v-col cols="6">
                        <v-text-field
                          v-model="editedItem.eligibility"
                          type="number"
                          label="Eligibility"
                          :rules="eligibilityRules"
                          required
                        ></v-text-field>
                      </v-col>
                    </v-row>
                  </v-container>
                </v-card-text>

                <v-card-actions>
                  <v-spacer></v-spacer>
                  <v-btn color="red darken-1" text @click="onClickCancel"
                    >Cancel</v-btn
                  >
                  <v-btn
                    color="green darken-1"
                    text
                    @click="onClickSave"
                    :disabled="!validInput"
                    >Save</v-btn
                  >
                </v-card-actions>
              </v-card>
            </v-form>
          </v-dialog>
        </td>
      </tr>
    </template>
  </v-data-table>
</template>

<script>
export default {
  props: ["participants", "initialEligibilities"],

  data: () => ({
    dialog: false,
    headers: [
      { text: "Name", value: "displayName" },
      { text: "MSP ID", value: "principal.mspId" },
      { text: "DN", value: "principal.dn", sortable: false },
      { text: "Eligibility", value: "eligibility", sortable: false },
      { text: "Actions", value: "action", sortable: false, align: "right" }
    ],
    nameRules: [v => !!v || "Name is required"],
    mspRules: [v => !!v || "MSP ID is required"],
    dnRules: [v => !!v || "DN is required"],
    eligibilityRules: [v => v >= 0 || "Must be a positive value"],

    validInput: false,
    editedIndex: -1,
    editedItem: {},
    defaultItem: {
      id: "",
      displayName: "",
      principal: {
        mspid: "",
        dn: ""
      },
      eligibility: 0
    }
  }),

  created() {
    this.editedItem = JSON.parse(JSON.stringify(this.defaultItem));
  },

  computed: {
    formTitle() {
      return this.editedIndex === -1 ? "New Participant" : "Edit Participant";
    },

    mergedData() {
      return this.participants.map(bidder => {
        const container = {};
        Object.assign(container, bidder);
        Object.assign(
          container,
          this.initialEligibilities
            .filter(y => y.bidderId === bidder.id)
            .map(el => ({ eligibility: el.number }))
            .shift()
        );
        return container;
      });
    }
  },

  watch: {
    dialog(val) {
      val || this.onClickCancel();
    }
  },

  methods: {
    onClickEditItem(item) {
      this.editedIndex = this.mergedData.indexOf(item);
      this.editedItem = JSON.parse(JSON.stringify(item));
      this.$refs.form.resetValidation();
      this.dialog = true;
    },

    onClickDeleteItem(item) {
      const pIndex = this.participants.indexOf(item);
      const eIndex = this.initialEligibilities.findIndex(
        ({ bidderId }) => bidderId === item.id
      );

      // remove
      this.participants.splice(pIndex, 1);
      this.initialEligibilities.splice(eIndex, 1);
    },

    onClickCancel() {
      this.dialog = false;
      setTimeout(() => {
        this.$refs.form.reset();
        this.editedItem = JSON.parse(JSON.stringify(this.defaultItem));
        this.editedIndex = -1;
      }, 300);
    },

    onClickSave() {
      let bidderId =
        this.editedIndex > -1
          ? this.editedItem.id
          : this.participants.length + 1;

      this.editedItem.id = bidderId;
      // get eligibility first
      let eligibility = {
        bidderId: bidderId,
        number: Number(this.editedItem.eligibility)
      };
      delete this.editedItem["eligibility"];

      if (this.editedIndex > -1) {
        this.participants[this.editedIndex] = JSON.parse(
          JSON.stringify(this.editedItem)
        );
        const eIndex = this.initialEligibilities.findIndex(
          ({ bidderId }) => bidderId === eligibility.bidderId
        );
        Object.assign(this.initialEligibilities[eIndex], eligibility);
      } else {
        this.participants.push(this.editedItem);
        this.initialEligibilities.push(eligibility);
      }
      this.emitUpdate();
      this.onClickCancel();
    },

    emitUpdate() {
      this.$emit(
        "update-participants",
        this.participants,
        this.initialEligibilities
      );
    }
  }
};
</script>
