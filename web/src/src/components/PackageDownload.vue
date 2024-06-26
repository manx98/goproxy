<script setup>
import {watch, reactive, ref, shallowRef, onMounted, onUnmounted} from "vue";
import {WarpFetch} from "~/utils";
import {ElMessage} from "element-plus";

let downloading = shallowRef(false);
let depName = shallowRef("");
let msgList = reactive([]);
let showModInput = shallowRef(false);

function convertToHtml(isErr, content) {
  content = content.replace("\n", "<br>")
  if (isErr) {
    return `<span style="color: red;font-size: 10px">${content}</span>`;
  } else {
    return `<span style="color: white;font-size: 10px">${content}</span>`;
  }
}

async function download() {
  if (!depName.value) {
    ElMessage.error("请输入依赖名称");
    return
  }
  showModInput.value = false;
  try {
    downloading.value = true;
    let response = await WarpFetch(fetch('/api/download_mod', {
      method: 'POST',
      body: depName.value
    }));
    msgList.length = 0;
    msgList.push(convertToHtml(false, '开始下载...\n'))
    response.onBinaryWrite = (val) => {
      if (val.length > 0) {
        let code = val[val.length - 1];
        switch (code) {
          case 69:
            let info_txt = new TextDecoder().decode(val.slice(0, val.length - 1));
            msgList.push(convertToHtml(true, info_txt));
            break;
          case 79:
            let err_txt = new TextDecoder().decode(val.slice(0, val.length - 1));
            msgList.push(convertToHtml(false, err_txt));
            break;
          default:
            throw new Error('未知状态码: ' + code)
        }
      }
    }
    response.onStatusUpdated = (val) => {
      if (val) {
        throw new Error(val)
      } else {
        msgList.push(convertToHtml(false, "\n下载完成"))
      }
    }
    await response.run();
  } catch (e) {
    msgList.push(convertToHtml(true, "\n发生异常: " + e))
  } finally {
    downloading.value = false;
  }
}

let msgBox = ref();

watch(msgList, () => {
  setTimeout(() => {
    msgBox.value.$el.scrollTop = msgBox.value.$el.scrollHeight;
  }, 0)
})

let terminalHeight = shallowRef('0px');

function changeTerminalHeight() {
  terminalHeight.value = window.innerHeight - 104 + 'px';
}

onMounted(changeTerminalHeight);

addEventListener("resize", changeTerminalHeight);

onUnmounted(() => {
  removeEventListener("resize", changeTerminalHeight);
});

function doShowModInput() {
  showModInput.value = true;
  if (!depName.value) {
    depName.value = `module mod_download
go 1.18.0
require (

)`
  }
}
</script>

<template>
  <el-dialog :close-on-click-modal="false" v-model="showModInput" title="输入mod文件内容">
    <el-input v-model="depName" type="textarea" :rows="15"></el-input>
    <div style="text-align: right">
      <el-button @click="download" type="primary">确认</el-button>
    </div>
  </el-dialog>
  <div style="text-align: center">
    <el-button @click="doShowModInput" :loading="downloading">{{ downloading ? "下载中" : "下载" }}</el-button>
  </div>
  <el-card :style="{'background-color': '#0a0a0a','height': terminalHeight, 'overflow-y': 'scroll'}" ref="msgBox">
    <span v-for="(item, index) in msgList" :key="index" v-html="item"></span>
  </el-card>
</template>

<style scoped>
</style>