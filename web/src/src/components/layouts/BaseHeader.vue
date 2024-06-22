<script setup>
import logoImg from '/public/logo.svg'
import {toRef} from "vue";

let pathname = toRef(window.location, "pathname")
</script>

<template>
  <el-menu class="el-menu-demo" mode="horizontal" text-color="withe" :default-active="pathname" :router="true">
    <el-menu-item index="/">
      <el-image :src="logoImg" class="logo"></el-image>
      GoProxy
    </el-menu-item>
  </el-menu>
</template>

<style scoped>
.logo {
  height: 45px;
  padding: 0 10px 0 0;
}

.ep-menu .ep-menu-item:hover,
.ep-menu .is-active {
  background-color: #fff !important;
}
</style>
