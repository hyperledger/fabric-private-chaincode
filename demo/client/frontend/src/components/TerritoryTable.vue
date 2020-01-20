<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-data-table
    :headers="headers"
    :items="territories"
    class="elevation-1"
    hide-default-footer
  >
    <template v-slot:top>
      <v-toolbar flat color="white">
        <v-toolbar-title>Territories</v-toolbar-title>
        <v-spacer />
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
                          v-model="editedItem.name"
                          label="Territory name"
                          :rules="nameRules"
                          required
                        />
                      </v-col>
                      <v-col cols="12" sm="6" md="4">
                        <v-text-field
                          v-model="editedItem.minPrice"
                          label="Min Price"
                          type="number"
                          :rules="minPriceRules"
                          required
                        />
                      </v-col>
                      <v-col cols="12">
                        <v-textarea
                          v-model="editedItem.channelsString"
                          name="channels"
                          label="Channels"
                          auto-grow
                          filled
                        />
                      </v-col>
                    </v-row>
                  </v-container>
                </v-card-text>

                <v-card-actions>
                  <v-spacer />
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
  props: ["territories"],
  data: () => ({
    dialog: false,
    headers: [
      { text: "Name", value: "name" },
      { text: "Min Price", value: "minPrice" },
      { text: "Channel Impairment", value: "channels", sortable: false },
      { text: "Actions", value: "action", sortable: false, align: "right" }
    ],
    validInput: true,
    minPriceRules: [v => v >= 0 || "Must be a positive value"],
    editedIndex: -1,
    editedItem: {},
    defaultItem: {
      id: "",
      name: "Some Name",
      minPrice: 10000,
      channelsString: '[{ "id": 1, "name": "a", "impairment": 0 }]',
      channels: []
    }
  }),

  created() {
    this.editedItem = this.defaultItem;
  },

  computed: {
    formTitle() {
      return this.editedIndex === -1 ? "New Territory" : "Edit Territory";
    },

    nameRules() {
      return [
        v => !!v || "Name is required"
        // v => (!!v && !this.territories.some(x => x.name === v)) || "Name already exists",
      ];
    }
  },

  watch: {
    dialog(val) {
      val || this.onClickCancel();
    }
  },

  methods: {
    onClickEditItem(item) {
      this.editedIndex = this.territories.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.editedItem.channelsString = JSON.stringify(this.editedItem.channels);
      this.dialog = true;
    },

    onClickDeleteItem(item) {
      const index = this.territories.indexOf(item);
      this.territories.splice(index, 1);
      this.emitUpdate();
    },

    onClickCancel() {
      this.dialog = false;
      setTimeout(() => {
        this.editedItem = JSON.parse(JSON.stringify(this.defaultItem));
        this.editedIndex = -1;
      }, 300);
    },

    onClickSave() {
      this.editedItem.id =
        this.editedIndex > -1
          ? this.editedItem.id
          : this.territories.length + 1;
      this.editedItem.channels = JSON.parse(this.editedItem.channelsString);
      if (this.editedIndex > -1) {
        Object.assign(this.territories[this.editedIndex], this.editedItem);
      } else {
        this.territories.push(this.editedItem);
      }
      this.emitUpdate();
      this.onClickCancel();
    },

    emitUpdate() {
      this.$emit("update-territories", this.territories);
    }
  }
};
</script>
