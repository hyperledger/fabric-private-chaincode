<template>
  <v-card
      class="mx-auto"
      max-width="400">
    <v-card-title class="headline font-weight-regular green white--text">
      <v-icon class="mr-2" large color="white">mdi-database-arrow-up</v-icon>
      <span class="headline">Upload Data</span>
    </v-card-title>
    <v-card-text>

      <v-row class="mx-auto">
        <v-col
            class="text-center" cols="12">
          <div class="container" v-cloak @drop.prevent="addFile" @dragover.prevent>
            <h2>File to Upload</h2>
            <div
                class="imagePreviewWrapper"
                :style="{ 'background-image': `url(${previewImage})` }">
              <h3 v-if="previewImage==null">(Drag it over)</h3>
            </div>
            <span v-if="file">{{ file.name }} ({{ file.size | kb }} kb)</span>
          </div>
        </v-col>
      </v-row>
      <v-row class="mx-auto">
        <v-col>
          <v-select
              v-model="selectedDataDomain"
              :items="dataDomainItems"
              label="Data Domain"
              dense
          ></v-select>
          <v-combobox
              v-model="checkedAllowedUse"
              :items="allowedUseItems"
              label="Allowed Usage"
              multiple
              outlined
              dense
          ></v-combobox>
          <v-checkbox
              v-model="confirmed"
              :label="`I consent to use my personal data under the above defined conditions.`"
          ></v-checkbox>
        </v-col>
      </v-row>

    </v-card-text>
    <v-divider></v-divider>
    <v-card-actions>
      <v-spacer></v-spacer>
      <v-btn
          :disabled="uploadDisabled"
          class="ma-1"
          @click="upload"
      >Upload
      </v-btn>
    </v-card-actions>

    <UploadProgressStatus :dialog="openDialog" v-on:done="onDialogDone"/>
  </v-card>
</template>

<script>
import UploadProgressStatus from "@/components/Upload.vue"
import axios from 'axios';

export default {
  name: 'Consent',
  components: {
    UploadProgressStatus
  },
  props: {
    msg: String
  },
  data: () => ({
    checkedAllowedUse: [],
    allowedUseItems: [
      'Private',
      'Public',
      'Government',
    ],
    selectedDataDomain: '',
    dataDomainItems: [
      'Health',
      'Financial',
    ],
    confirmed: false,
    previewImage: null,
    imageData: null,
    file: null,
    openDialog: false
  }),
  computed: {
    uploadDisabled() {
      return this.file == null ||
          this.selectedDataDomain === '' ||
          this.checkedAllowedUse.length === 0 ||
          this.confirmed === false;
    }
  },
  methods: {
    addFile(e) {
      let droppedFiles = e.dataTransfer.files;
      if (!droppedFiles) {
        return;
      }
      // this tip, convert FileList to array, credit: https://www.smashingmagazine.com/2018/01/drag-drop-file-uploader-vanilla-js/
      ([...droppedFiles]).forEach(f => {
        console.log(f.name);
        this.file = f;

        let reader = new FileReader;
        reader.onload = e => {
          this.previewImage = e.target.result;
          this.imageData = reader.result
              .replace('data:', '')
              .replace(/^.+,/, '');
        };
        reader.readAsDataURL(f);

      });
    },
    upload() {
      console.log('upload');
      const BASE_URL = process.env.VUE_APP_BASEURL_DATA_PROVIDER
      axios.post(BASE_URL + '/api/upload', {
        data: this.imageData,
        dataName: this.file.name,
        domain: this.selectedDataDomain,
        allowedUse: this.checkedAllowedUse,
      }).then(response => {
        this.openDialog = true
        console.log(response);
      });
    },
    resetForm() {
      console.log('reset form');
      this.file = null;
      this.checkedAllowedUse = [];
      this.selectedDataDomain = '';
      this.confirmed = false;
      this.previewImage = null;
      this.imageData = null;
    },
    onDialogDone() {
      this.resetForm();
      this.openDialog = false
    }
  },
  filters: {
    kb(val) {
      return Math.floor(val / 1024);
    }
  },
};

</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>

.container {
  width: 350px;
  height: 350px;
  border: 2px dotted gray;
  margin: 10px auto;
}

.imagePreviewWrapper {
  width: 250px;
  height: 250px;
  display: block;
  cursor: pointer;
  margin: 10px auto 10px;
  background-size: cover;
  background-position: center center;
}
</style>
