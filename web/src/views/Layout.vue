<template>
  <div class="layout">
    <aside class="sidebar">
      <h3>V2Wall</h3>
      <router-link to="/ips">📊 IP 统计</router-link>
      <router-link to="/whitelist">📋 白名单</router-link>
      <router-link to="/config">⚙️ 配置</router-link>
      <router-link to="/nodes">🖥️ 节点</router-link>
      <a @click.prevent="logout">🚪 登出</a>
    </aside>
    <main class="content">
      <router-view />
    </main>
  </div>
</template>

<script setup>
import { useRouter } from 'vue-router'

const router = useRouter()
const logout = () => {
  localStorage.removeItem('token')
  router.push('/login')
}
</script>

<style scoped>
.layout {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 200px;
  background: #304156;
  color: white;
  padding: 20px 0;
}
.sidebar h3 {
  text-align: center;
  margin-bottom: 30px;
}
.sidebar a {
  display: block;
  padding: 12px 20px;
  color: #bfcbd9;
  text-decoration: none;
}
.sidebar a:hover, .sidebar a.router-link-active {
  color: white;
  background: #263445;
}
.content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: #f0f2f5;
}
</style>
