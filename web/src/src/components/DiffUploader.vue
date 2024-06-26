<script setup>
import tailSpin from "~/assets/icon/tail-spin.svg";
import {DelayNumUpdater, formatBytes, WarpFetch} from "~/utils";
import {shallowRef} from "vue";
import {DelayUpdateMs} from "~/config";

let msg = shallowRef("");
let running = shallowRef(false);
let stepNum = shallowRef(0);
let dirTotal = shallowRef(0);
let dirNum = new DelayNumUpdater(DelayUpdateMs);
let totalSize = shallowRef(0);
let fileSize = new DelayNumUpdater(DelayUpdateMs);

async function uploadDownload(file) {
  try {
    dirNum.reset();
    fileSize.reset();
    stepNum.value = 1;
    totalSize.value = file.size;
    msg.value = "";
    running.value = true;
    const formData = new FormData();
    formData.append('file', file);
    let response = await WarpFetch(fetch('/api/apply_diff', {
      method: 'POST',
      body: formData,
      upload: (progressEvent) => {
        fileSize.setValue(progressEvent.loaded);
      }
    }));
    response.onDirAddUpdated = (val)=>{
      dirNum.add(val);
    };
    response.onSizeAddUpdated = (val)=>{
      fileSize.add(val);
    };
    response.onBinaryWrite = (val)=>{
      let data = JSON.parse(new TextDecoder().decode(val));
      dirTotal.value = data.num;
      totalSize.value = data.total_size;
      stepNum.value = 2;
    };
    await response.run();
    stepNum.value = 3;
  } catch(e){
    stepNum.value = 4;
    msg.value = "上传失败:" +e;
  }finally {
    running.value = false;
    dirNum.flush();
    fileSize.flush();
  }
}
function beforeUpload(file) {
  uploadDownload(file);
  return false;
}

</script>

<template>
  <div>
    <el-upload :disabled="running" :before-upload="beforeUpload">
      <el-image v-show="running" :src="tailSpin" style="width: 20px"></el-image>
      上传差异文件
    </el-upload>
    <div style="font-size: 12px">
    <span v-if="stepNum === 1">
      正在上传差异文件，请稍后({{ formatBytes(totalSize) }}/{{ formatBytes(fileSize.Value.value) }})
    </span>
      <span v-if="stepNum === 2">
      正在解压数据库，请稍后({{ formatBytes(totalSize) }}/{{ formatBytes(fileSize.Value.value) }}),剩余 {{ dirTotal - dirNum.Value.value }} 待解压
    </span>
    <span v-if="stepNum === 3" style="color: green">
      上传成功.
    </span>
    <span v-if="stepNum === 4" style="color: red">
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