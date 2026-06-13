<template>
  <div>
    <h2>拉黑配置</h2>
    <div class="form">
      <label>保留天数：<input v-model.number="form.retention_days" type="number" /></label>
      <label>同步间隔（秒）：<input v-model.number="form.sync_interval_sec" type="number" /></label>
      <button @click="save">保存</button>
    </div>
    <p>最后同步时间：{{ lastSyncStr }}</p>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { getConfig, updateConfig } from '../api'

const form = ref({
  retention_days: 30,
  sync_interval_sec: 60,
  last_sync_time: 0
})

const lastSyncStr = computed(() => {
  const ts = form.value.last_sync_time
  if (!ts) return '从未同步'
  return new Date(ts / 1000000).toLocaleString() // 纳秒转毫秒
})

onMounted(async () => {
  const res = await getConfig()
  Object.assign(form.value, res.data)
})

const save = async () => {
  await updateConfig({
    retention_days: form.value.retention_days,
    sync_interval_sec: form.value.sync_interval_sec
  })
  alert('保存成功')
}
</script>
