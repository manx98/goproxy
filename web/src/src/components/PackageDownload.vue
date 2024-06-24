<script setup>
import {watch, reactive, ref, shallowRef} from "vue";
import {Fetch} from "~/utils";
import qs from 'qs';

let downloading = shallowRef(false);
let depName = shallowRef("");
let msgList = reactive([]);

function convertToHtml(isErr, content) {
  content = content.replace("\\n", "<br>")
  if (isErr) {
    return `<span style="color: red;font-size: 10px">${content}</span>`;
  } else {
    return `<span style="color: white;font-size: 10px">${content}</span>`;
  }
}

async function download() {
  try {
    downloading.value = true;
    let response = await Fetch('/api/download_mod?' + qs.stringify({
      "q": depName.value,
      "v": version.value,
    }), 'GET');
    msgList.length = 0;
    msgList.push(convertToHtml(false, '开始下载: ' + depName.value))
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
        msgList.push(convertToHtml(false, "下载完成"))
      }
    }
    await response.run();
  } catch (e) {
    msgList.push(convertToHtml(true, "发生异常: " + e))
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
</script>

<template>
  <el-row :gutter="10">
    <el-col :span="14">
      <el-input :disabled="downloading" v-model="depName" placeholder="输入包名(golang.org/x/mod)"></el-input>
    </el-col>
    <el-col :span="10">
      <el-input v-model="version" :disabled="downloading" placeholder="输入版本号(v0.17.0)">
        <template #append>
          <el-button @click="download" :loading="downloading">{{ downloading ? "下载中" : "下载" }}</el-button>
        </template>
      </el-input>
    </el-col>
  </el-row>
  <el-card style="background-color: #0a0a0a;height: 30vh;overflow-y: scroll" ref="msgBox">
    <p v-for="(item, index) in msgList" :key="index" v-html="item"></p>
  </el-card>
</template>

<style scoped>
</style>