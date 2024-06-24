<script setup>
import {shallowRef} from "vue";
import {Fetch, formatBytes, formatTime, getDiffZipFileName} from '~/utils';
import StreamSaver from "streamsaver";
import tailSpin from "~/assets/icon/tail-spin.svg";
import qs from "qs";

defineProps({
  info: Object
})

let show = shallowRef(false);
let msg = shallowRef("");
let occurErr = shallowRef(false);
let running = shallowRef(false);
let dirNum = shallowRef(0);
let fileSize = shallowRef(0);

async function downloadDiff(id) {
  let fileStream = null;
  msg.value = "正在下载,请勿关闭页面";
  occurErr.value = false;
  dirNum.value = 0;
  fileSize.value = 0;
  running.value = true;
  show.value = true;
  try {
    fileStream = StreamSaver.createWriteStream(getDiffZipFileName()).getWriter();
    let resp = await Fetch("/api/download_diff?" + qs.stringify({id: id}), 'GET');
    resp.onBinaryWrite = async function (val) {
      await fileStream.write(val);
    };
    resp.onStatusUpdated = (val) => {
      if (val) {
        throw new Error(val)
      }
    };
    resp.onDirAddUpdated = (val) => {
      dirNum.value += val;
    };
    resp.onSizeAddUpdated = (val) => {
      fileSize.value += val;
    };
    await resp.run();
    await fileStream.close();
    msg.value = "下载完成";
  } catch (e) {
    if (fileStream) {
      await fileStream.abort()
    }
    occurErr.value = true;
    msg.value = "下载出错: " + e;
  } finally {
    running.value = false;
  }
}
</script>

<template>
  <el-card>
    <el-space v-if="info.id">
      <el-button type="text" @click="downloadDiff(info.id)">下载</el-button>
      <el-tag type="primary">{{ formatTime(info.mtime) }}</el-tag>
      <el-tag type="info">{{ info.desc || '无任何描述' }}</el-tag>
    </el-space>
    <el-space v-else>
      <el-button @click="downloadDiff(info.id)">整库下载</el-button>
      <el-upload>上传差异文件</el-upload>
    </el-space>
  </el-card>
  <el-dialog v-model="show" :show-close="!running" :close-on-click-modal="!running">
    <table>
      <tr>
        <td>已遍历(目录)</td>
        <td>{{ dirNum }}</td>
      </tr>
      <tr>
        <td>已压缩</td>
        <td>{{ formatBytes(fileSize) }}</td>
      </tr>
      <tr>
        <td>状态</td>
        <td v-if="running">
          <el-image :src="tailSpin"></el-image>
          {{ msg }}
        </td>
        <td v-else>
          <span v-if="occurErr" style="color: red">
            {{ msg }}
          </span>
          <span v-else>
            {{ msg }}
          </span>
        </td>
      </tr>
    </table>
  </el-dialog>
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