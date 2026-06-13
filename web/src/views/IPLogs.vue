<template>
  <div>
    <h2>查询日志</h2>
    <input v-model="searchIP" placeholder="输入 IP" />
    <button @click="search">查询</button>
    <div v-if="logs.length">
      <table>
        <thead><tr><th>时间</th><th>方法</th><th>URL</th><th>UA</th><th>Referer</th></tr></thead>
        <tbody>
          <tr v-for="log in logs" :key="log.time">
            <td>{{ log.time_str }}</td>
            <td>{{ log.method }}</td>
            <td>{{ log.url }}</td>
            <td>{{ log.user_agent }}</td>
            <td>{{ log.referer }}</td>
          </tr>
        </tbody>
      </table>
      <Pagination :total="total" :page="page" :size="size" @change="doSearch" />
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { getIPLogs } from '../api'
import Pagination from '../components/Pagination.vue'

const searchIP = ref('')
const logs = ref([])
const page = ref(1)
const size = ref(20)
const total = ref(0)

const search = () => {
  page.value = 1
  doSearch()
}

const doSearch = async () => {
  try {
    const res = await getIPLogs(searchIP.value, page.value, size.value)
    logs.value = res.data.data
    total.value = res.data.total
  } catch (e) {
    console.error(e)
  }
}
</script>
