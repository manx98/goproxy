<template>
  <el-input v-model="desc" placeholder="检查点描述">
    <template #append>
      <el-button @click="createCheckpoint">创建检查点</el-button>
    </template>
  </el-input>
  <el-dialog v-model="show" :show-close="!running" :close-on-click-modal="!running">
    <table>
      <tr>
        <td>目录数</td>
        <td>{{ dirNum }}</td>
      </tr>
      <tr>
        <td>文件数</td>
        <td>{{ fileNum }}</td>
      </tr>
      <tr>
        <td>大小</td>
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

<script setup>
import tailSpin from '~/assets/icon/tail-spin.svg'
import {Fetch, formatBytes} from '~/utils'
import {shallowRef} from 'vue'
import qs from 'qs'

let show = shallowRef(false);
let desc = shallowRef("");
let msg = shallowRef("");
let occurErr = shallowRef(false);
let running = shallowRef(false);
let dirNum = shallowRef(0);
let fileNum = shallowRef(0);
let fileSize = shallowRef(0);

async function createCheckpoint() {
  try {
    running.value = true;
    msg.value = "正在创建检查点..."
    dirNum.value = 0;
    fileNum.value = 0;
    fileSize.value = 0;
    occurErr.value = false;
    show.value = true;
    const response = await Fetch('/api/create_checkpoint?' + qs.stringify({
      desc: desc.value,
    }), 'GET');
    response.onBinaryWrite = (val) => {
      if(val.length) {
        msg.value = "已生成新版本."
      } else {
        msg.value = "没有生成新版本."
      }
    }
    response.onDirAddUpdated = (val) => {
      dirNum.value += val;
    }
    response.onSizeAddUpdated = (val) => {
      fileNum.value += 1;
      fileSize.value += val;
    }
    response.onStatusUpdated = (val) => {
      if(val) {
        occurErr.value = true;
        msg.value = "发生错误: " + val;
      }
    }
    await response.run()
  } catch (error) {
    occurErr.value = true;
    msg.value = "发生错误: " + error;
  } finally {
    running.value = false;
  }
}
</script>

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