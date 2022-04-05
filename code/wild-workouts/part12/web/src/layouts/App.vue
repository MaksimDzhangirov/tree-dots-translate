<template>
  <div>
    <WebMenu/>
      <div class="container">
        <slot></slot>
        <WebFooter/>
      </div>
  </div>
</template>

<script>
import WebFooter from "../components/WebFooter";
import WebMenu from "../components/WebMenu";
import {Auth} from "../repositories/auth";

export default {
  name: "Main",
  components: {
    WebFooter,
    WebMenu,
  },
  beforeCreate: function () {
    document.body.className = 'app';

    if (!Auth.isLoggedIn()) {
      this.$router.push({name: 'login'})
    }
  }
}
</script>

<style scoped>

</style>