<script setup>
import tailSpin from "~/assets/icon/tail-spin.svg";
import {DelayNumUpdater, formatBytes} from "~/utils";
import {shallowRef} from "vue";
import {DelayUpdateMs} from "~/config";

let msg = shallowRef("");
let running = shallowRef(false);
let stepNum = shallowRef(0);
let dirTotal = shallowRef(0);
let dirNum = new DelayNumUpdater(DelayUpdateMs);
let totalSize = shallowRef(0);
let fileSize = new DelayNumUpdater(DelayUpdateMs);

function beforeUpload(file) {
  return false;
}

</script>

<template>
  <div>
    <el-upload :disabled="running" :before-upload="beforeUpload">
      <el-image v-show="running" :src="tailSpin" style="width: 20px"></el-image>
      上传差异文件
    </el-upload>
    <div style="font-size: 14px">
    <span v-if="stepNum === 1">
      正在上传差异文件，请稍后({{ formatBytes(totalSize) }}/{{ formatBytes(fileSize.Value.value) }})
    </span>
      <span v-if="stepNum === 2">
      正在解压数据库，请稍后({{ formatBytes(totalSize) }}/{{ formatBytes(fileSize.Value.value) }}),剩余 {{ dirTotal - dirNum.Value.value }} 待解压
    </span>
      <span v-if="stepNum === 3" style="color: red">
      操作失败: {{ msg }}
    </span>
    </div>
  </div>
</template>

<style scoped>
table {
  width: 100%;
  border-collapse: collapse;
}

table caption {
  font-size: 2em;
  font-weight: bold;
  margin: 1em 0;
}

th, td {
  border: 1px solid #999;
  text-align: center;
  padding: 20px 0;
}

table thead tr {
  background-color: #008c8c;
  color: #fff;
}

table tbody tr:nth-child(odd) {
  background-color: #eee;
}

table tbody tr:hover {
  background-color: #ccc;
}

table tbody tr td:first-child {
  color: #f40;
}

table tfoot tr td {
  text-align: right;
  padding-right: 20px;
}
</style>