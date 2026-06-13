<template>
  <div>
    <h2>被控节点</h2>
    <table>
      <thead><tr><th>节点ID</th><th>最后心跳</th><th>诱饵开启</th><th>最后上报时间</th></tr></thead>
      <tbody>
        <tr v-for="node in nodes" :key="node.id">
          <td>{{ node.id }}</td>
          <td>{{ formatTs(node.last_heartbeat) }}</td>
          <td>{{ node.honeypot_enabled ? '是' : '否' }}</td>
          <td>{{ formatTs(node.last_report_time) }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getNodes } from '../api'

const nodes = ref([])

const formatTs = (ts) => {
  if (!ts) return '无'
  return new Date(ts / 1000000).toLocaleString()
}

onMounted(async () => {
  const res = await getNodes()
  nodes.value = res.data.data
})
</script>
