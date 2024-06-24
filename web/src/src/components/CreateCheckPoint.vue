<template>
  <el-input v-model="desc" placeholder="检查点描述">
    <template #append>
      <el-button @click="createCheckpoint">创建检查点</el-button>
    </template>
  </el-input>
  <div :style="{height: containerBoxHeight, overflow: 'auto'}">
    <div
        v-infinite-scroll="loadHead"
        :infinite-scroll-disabled="disabled"
    >
      <DiffInfo v-for="item in data" :key="item.id" :info="item"></DiffInfo>
    </div>
    <div v-if="loading" style="text-align: center">
      <el-image :src="tailSpin"></el-image>
    </div>
  </div>
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
import {computed, onMounted, onUnmounted, reactive, shallowRef, watch} from 'vue'
import qs from 'qs'
import {ElMessage} from "element-plus";

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
    if(!desc.value) {
      ElMessage.error("请输入检查点描述.");
      return
    }
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
      if (val.length) {
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
      if (val) {
        occurErr.value = true;
        msg.value = "发生错误: " + val;
      }
    }
    await response.run()
    desc.value = "";
  } catch (error) {
    occurErr.value = true;
    msg.value = "发生错误: " + error;
  } finally {
    running.value = false;
  }
}

const loading = shallowRef(false);
const noMore = shallowRef(true);
const disabled = computed(() => loading.value || noMore.value)
const data = reactive([]);
let lastId = "";

watch(show, (v) => {
  if (!v) {
    reloadHead();
  }
})

function reloadHead() {
  lastId = "";
  data.length = 0;
  data.push({
    id: "",
    mtime: new Date().getTime(),
    desc: "",
  })
  noMore.value = false;
  loading.value = false;
}

async function loadHead() {
  try {
    loading.value = true;
    let result = await (await fetch('/api/get_header?' + qs.stringify({
      id: lastId,
      num: 10,
    }))).json();
    if (!result.data) {
      noMore.value = true
    } else {
      result.data.forEach((v) => {
        data.push(v);
        lastId = v.id;
      });
      noMore.value = result.data.length < 10;
    }
  } catch (e) {
    noMore.value = true;
    alert(e.toString())
  } finally {
    loading.value = false;
  }
}

let containerBoxHeight = shallowRef('0px');

function changeContainerBoxHeight() {
  containerBoxHeight.value = window.innerHeight - 104 + 'px';
}

onMounted(() => {
  changeContainerBoxHeight();
  reloadHead();
});

addEventListener("resize", changeContainerBoxHeight);

onUnmounted(() => {
  removeEventListener("resize", changeContainerBoxHeight);
});
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