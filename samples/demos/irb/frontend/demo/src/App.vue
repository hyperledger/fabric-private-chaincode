<template>
  <v-app>
    <AppNavi></AppNavi>

    <!-- Sizes your content based upon application components -->
    <v-main>

      <!-- Provides the application the proper gutter -->
      <v-container fluid>

        <!-- If using vue-router -->
        <router-view></router-view>
      </v-container>
    </v-main>

    <v-footer app>
      <!-- -->
    </v-footer>
  </v-app>


</template>

<script>
import AppNavi from "@/components/layout/AppNavi.vue"

export default {
  components: {
    AppNavi
  }
}
</script>

<style lang="scss">

</style>
