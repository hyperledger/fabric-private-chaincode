<!--
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
-->

<template>
  <v-card>
    <v-card-text>
      <!--      <v-alert prominent type="warning" class="mt-4">-->
      <!--        <div>-->
      <!--          Warning! This view can manipulate the peer's local ledger state.-->
      <!--        </div>-->
      <!--        <div>-->
      <!--          Any actions performed here may result in invalid chaincode execution-->
      <!--          results.-->
      <!--        </div>-->
      <!--      </v-alert>-->

      <v-data-table
        :headers="headers"
        :items="state"
        class="elevation-0"
        hide-default-footer
        :loading="isLoading"
        :calculate-widths="true"
        loading-text="Loading... Please wait"
      >
        <template v-slot:item.value="props">
          {{ props.value | jsonEvenPrettier | trim }}
        </template>

        <template v-slot:header.name="{ header }">
          {{ header.text.toUpperCase() }}
        </template>

        <template v-slot:item.action="{ item }">
          <div>
            <v-tooltip top>
              <template v-slot:activator="{ on }">
                <v-icon
                  small
                  class="mr-2"
                  v-on="on"
                  @click="onClickEditItem(item)"
                >
                  fa-edit
                </v-icon>
              </template>
              <span>Edit</span>
            </v-tooltip>

            <v-tooltip top>
              <template v-slot:activator="{ on }">
                <v-icon small v-on="on" @click="onClickDeleteItem(item)">
                  fa-trash
                </v-icon>
              </template>
              <span>Delete</span>
            </v-tooltip>
          </div>
        </template>

        <template v-slot:body.append="{ headers }">
          <tr>
            <td :colspan="headers.length" class="text-right">
              <v-dialog v-model="dialog" max-width="600px">
                <template v-slot:activator="{ on }">
                  <v-btn icon v-on="on"><v-icon>fa-plus-circle</v-icon></v-btn>
                </template>
                <v-form ref="form" v-model="validInput">
                  <v-card>
                    <v-card-title
                      ><span class="headline">{{
                        formTitle
                      }}</span></v-card-title
                    >
                    <v-card-text>
                      <v-container>
                        <v-row>
                          <v-col cols="12">
                            <v-text-field
                              v-model="editedItem.key"
                              label="Key"
                              :rules="keyRules"
                              required
                              :disabled="editedIndex !== -1"
                            />
                          </v-col>
                          <v-col cols="12">
                            <v-textarea
                              v-model="editedItem.value"
                              name="value"
                              label="Value"
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
    </v-card-text>
  </v-card>
</template>

<script>
import { mapState, mapActions } from "vuex";

export default {
  data: () => ({
    dialog: false,
    isLoading: true,
    headers: [
      { text: "Key", value: "key", width: "20px" },
      { text: "Value", value: "value", width: "60px" },
      {
        text: "Actions",
        value: "action",
        sortable: false,
        align: "right",
        width: "100px"
      }
    ],
    validInput: true,
    editedIndex: -1,
    editedItem: {},
    defaultItem: {
      key: "",
      value: ""
    }
  }),

  created() {
    this.editedItem = this.defaultItem;
  },

  computed: {
    ...mapState({
      state: state => state.ledger.state
    }),

    formTitle() {
      return this.editedIndex === -1 ? "New Entry" : "Edit Entry";
    },

    keyRules() {
      return [v => !!v || "Key is required"];
    }
  },

  watch: {
    dialog(val) {
      val || this.onClickCancel();
    }
  },

  mounted() {
    this.fetchState()
      .catch(err => console.log("error: " + err))
      .finally(() => (this.isLoading = false));
  },

  filters: {
    jsonEvenPrettier: function(value) {
      if (!value) return "";
      value = value.toString();
      try {
        value = JSON.stringify(JSON.parse(value));
      } catch (e) {
        // if this is not a json ... we just dont care
      }
      return value;
    },
    trim: function(value) {
      if (!value) return "";
      value = value.toString();
      if (value.length > 60) {
        return value.substring(0, 60) + " ...";
      }
      return value;
    }
  },

  methods: {
    ...mapActions({
      fetchState: "ledger/fetchState",
      deleteStateItem: "ledger/deleteStateItem",
      updateStateItem: "ledger/updateStateItem"
    }),

    onClickEditItem(item) {
      this.editedIndex = this.state.indexOf(item);
      this.editedItem = Object.assign({}, item);
      this.dialog = true;
    },

    onClickDeleteItem(item) {
      this.deleteStateItem(item).catch(err => console.log(err));
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
        this.editedIndex > -1 ? this.editedItem.id : this.state.length + 1;

      this.updateStateItem(this.editedItem)
        .catch(err => console.log(err))
        .finally(this.onClickCancel());
    }
  }
};
</script>
