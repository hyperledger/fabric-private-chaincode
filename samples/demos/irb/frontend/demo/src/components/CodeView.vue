<template>
  <v-container>
    <v-row class="mx-0 mt-0">
      <v-col cols="3" class="px-0">
        <v-icon>
          mdi-folder-open
        </v-icon>
        Files
        <v-treeview
            dense
            activatable
            open-on-click
            :items="files"
            @update:active="viewActiveFile"
        >
          <template v-slot:prepend="{ item, open }">
            <v-icon v-if="!item.file">
              {{ open ? 'mdi-folder-open' : 'mdi-folder' }}
            </v-icon>
            <v-icon v-else>
              {{ fileExtensions[item.file] }}
            </v-icon>
          </template>
        </v-treeview>
      </v-col>

      <v-col
          v-if="selectedFile === null"
          class="text-center"
      >
        <div v-if="!isReadOnly" class="drag-container align-center" v-cloak @drop.prevent="addFile" @dragover.prevent>
          <h2>Files to Upload</h2>
          <div>
            <h3>(Drag them over)</h3>
          </div>
        </div>
        <v-textarea
            v-if="codeIdentity !== ''"
            v-model="codeIdentity"
            label="Code Identity"
            outlined
            readonly
            rows="2"
            hide-details
        ></v-textarea>
      </v-col>

      <v-col
          cols="9"
          class="mx-0"
          v-else
      >
        <span>{{ selectedFile.name }}</span>
        <pre><code id="code-viewer" class="language-py">{{ selectedFile.content }}</code></pre>
      </v-col>
    </v-row>
  </v-container>
</template>

<script>

import Prism from 'prismjs';
import 'prismjs/components/prism-go';
import 'prismjs/components/prism-python';
import 'prismjs/themes/prism.css';
import 'prismjs/themes/prism-coy.css';

import {sha256} from 'js-sha256';

export default {
  name: 'CodeView',
  components: {},
  props: {
    files: Array,
    isReadOnly: Boolean,
  },
  mounted() {
    this.updateCodeIdentity();
    Prism.highlightAll();
  },
  data: () => ({
    codeIdentity: '',
    confirmed: false,
    selectedFile: null,
    fileExtensions: {
      md: 'mdi-language-markdown',
      go: 'mdi-language-go',
      py: 'mdi-language-python',
      json: 'mdi-code-json',
      txt: 'mdi-file-document-outline',
      manifest: 'mdi-file-document-outline',
    },
  }),
  computed: {
  },
  methods: {
    viewActiveFile(items) {
      if (items.length === 0) {
        this.selectedFile = null;
        return;
      }

      // note we only select a single item
      this.selectedFile = this.files.find(file => file.id === items[0]);
    },
    addFile(e) {
      if (this.isReadOnly) {
        return
      }

      let droppedFiles = e.dataTransfer.files;
      if (!droppedFiles) {
        return;
      }

      let alreadyUploaded = this.files.length;

      // this tip, convert FileList to array, credit: https://www.smashingmagazine.com/2018/01/drag-drop-file-uploader-vanilla-js/
      ([...droppedFiles]).forEach((f, i) => {
        console.log(f.name);

        let reader = new FileReader;
        reader.onload = e => {
          let fileContent = e.target.result;
          let extension = f.name.split(/(?:\.([^.]+))?$/)[1];
          let newFile = {id: alreadyUploaded + i, name: f.name, file: extension, content: fileContent};
          this.files.push(newFile);
          this.updateCodeIdentity();
        };
        reader.readAsText(f);

      });
    },
    updateCodeIdentity() {
      // update code identity, just som rubbish hashing ... don't do that!
      this.codeIdentity = this.files.reduce((acc, val) => sha256(acc + "||" + sha256(val.content)), "");
      this.$emit("update-files", this.codeIdentity)
    },
  },
  watch: {
    selectedFile: function (file) {
      if (file === null) {
        return;
      }
      let viewer = document.getElementById('code-viewer');
      if (viewer === null) {
        return;
      }
      viewer.innerHTML = file.content;
      this.$nextTick(() => Prism.highlightElement(viewer));
    }, deep: true
  },
  updated() {
    Prism.highlightAll();
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
.drag-container {
  height: 100px;
  border: 2px dotted gray;
  margin: 10px auto;
}
</style>
