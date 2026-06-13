<template>
  <div>
    <h2>白名单管理</h2>
    <div class="add-form">
      <input v-model="newCIDR" placeholder="CIDR (如 192.168.1.0/24)" />
      <input v-model="remark" placeholder="备注" />
      <button @click="add">添加</button>
    </div>
    <table>
      <thead><tr><th>CIDR</th><th>备注</th><th>操作</th></tr></thead>
      <tbody>
        <tr v-for="item in list" :key="item.cidr">
          <td>{{ item.cidr }}</td>
          <td>{{ item.remark }}</td>
          <td><button @click="del(item.cidr)">删除</button></td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { getWhitelist, addWhitelist, deleteWhitelist } from '../api'

const list = ref([])
const newCIDR = ref('')
const remark = ref('')

const fetch = async () => {
  const res = await getWhitelist()
  list.value = res.data.data
}

onMounted(fetch)

const add = async () => {
  await addWhitelist(newCIDR.value, remark.value)
  newCIDR.value = ''
  remark.value = ''
  await fetch()
}

const del = async (cidr) => {
  await deleteWhitelist(cidr)
  await fetch()
}
</script>
