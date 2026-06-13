<template>
  <div>
    <h2>IP 统计</h2>
    <table class="data-table">
      <thead>
        <tr>
          <th>IP</th>
          <th>地区</th>
          <th>访问次数</th>
          <th>最后访问时间</th>
          <th>操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="item in list" :key="item.ip">
          <td>{{ item.ip }}</td>
          <td>{{ item.info || '-' }}</td>
          <td>{{ item.count }}</td>
          <td>{{ item.last_seen_str }}</td>
          <td>
            <button @click="showLogs(item.ip)">查看日志</button>
          </td>
        </tr>
      </tbody>
    </table>
    <Pagination v-if="total" :total="total" :page="page" :size="size" @change="fetchData" />

    <!-- 日志弹窗 -->
    <div v-if="logVisible" class="modal">
      <div class="modal-content">
        <h3>访问日志 - {{ currentIP }}</h3>
        <div class="ipinfo" v-if="ipInfo">{{ ipInfo }}</div>
        <table>
          <thead><tr><th>时间</th><th>方法</th><th>URL</th><th>UA</th></tr></thead>
          <tbody>
            <tr v-for="log in logs" :key="log.time">
              <td>{{ log.time_str }}</td>
              <td>{{ log.method }}</td>
              <td>{{ log.url }}</td>
              <td>{{ log.user_agent }}</td>
            </tr>
          </tbody>
        </table>
        <Pagination :total="logTotal" :page="logPage" :size="logSize" @change="fetchLogs" />
        <button @click="logVisible = false">关闭</button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getIPList, getIPLogs, getIPInfo } from '../api'
import Pagination from '../components/Pagination.vue'

const list = ref([])
const page = ref(1)
const size = ref(20)
const total = ref(0)

const logVisible = ref(false)
const currentIP = ref('')
const logs = ref([])
const logPage = ref(1)
const logSize = ref(20)
const logTotal = ref(0)
const ipInfo = ref('')

onMounted(() => {
  fetchData()
})

const fetchData = async () => {
  try {
    const res = await getIPList(page.value, size.value)
    list.value = res.data.data
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  }
}

const showLogs = async (ip) => {
  currentIP.value = ip
  logPage.value = 1
  logVisible.value = true
  await fetchLogs()
  // 获取 ipinfo
  try {
    const res = await getIPInfo(ip)
    ipInfo.value = res.data.info
  } catch (e) {
    ipInfo.value = ''
  }
}

const fetchLogs = async () => {
  try {
    const res = await getIPLogs(currentIP.value, logPage.value, logSize.value)
    logs.value = res.data.data
    logTotal.value = res.data.total
  } catch (e) {
    console.error(e)
  }
}
</script>

<style scoped>
.data-table {
  width: 100%;
  border-collapse: collapse;
  background: white;
}
.data-table th, .data-table td {
  padding: 10px;
  border-bottom: 1px solid #ddd;
}
.modal {
  position: fixed;
  top: 0; left: 0;
  width: 100%; height: 100%;
  background: rgba(0,0,0,0.5);
  display: flex;
  justify-content: center;
  align-items: center;
}
.modal-content {
  background: white;
  padding: 20px;
  width: 80%;
  max-height: 80%;
  overflow-y: auto;
}
</style>
