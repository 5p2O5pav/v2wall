<template>
  <div class="pagination" v-if="total > size">
    <button :disabled="page <= 1" @click="$emit('change', page - 1)">上一页</button>
    <span>第 {{ page }} / {{ Math.ceil(total / size) }} 页 (共 {{ total }} 条)</span>
    <button :disabled="page >= Math.ceil(total / size)" @click="$emit('change', page + 1)">下一页</button>
  </div>
</template>

<script setup>
defineProps({
  total: Number,
  page: Number,
  size: Number
})
defineEmits(['change'])
</script>

<style scoped>
.pagination {
  margin: 15px 0;
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
