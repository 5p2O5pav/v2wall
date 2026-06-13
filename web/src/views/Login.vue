<template>
  <div class="login-container">
    <h2>V2Wall 管理登录</h2>
    <div class="login-form">
      <input v-model="username" placeholder="用户名" />
      <input v-model="password" type="password" placeholder="密码" />
      <button @click="doLogin">登录</button>
      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="needInit">尚未初始化，请输入初始Token进行初始化：</p>
      <input v-if="needInit" v-model="initToken" placeholder="初始Token" />
      <button v-if="needInit" @click="doInit">初始化并登录</button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { login, initAdmin } from '../api'

const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const initToken = ref('')
const error = ref('')
const needInit = ref(false)

onMounted(async () => {
  // 尝试调用 init 接口看是否需要初始化（返回403表示已初始化）
  try {
    await initAdmin('test', 'test', '') // 故意失败
  } catch (e) {
    if (e.response && e.response.status === 403) {
      // 已初始化
    } else if (e.response && e.response.status === 401) {
      needInit.value = true
    }
  }
})

const doLogin = async () => {
  try {
    const res = await login(username.value, password.value)
    localStorage.setItem('token', res.data.token)
    const redirect = route.query.redirect || '/'
    router.push(redirect)
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  }
}

const doInit = async () => {
  try {
    await initAdmin(username.value, password.value, initToken.value)
    // 初始化成功后自动登录
    const res = await login(username.value, password.value)
    localStorage.setItem('token', res.data.token)
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || '初始化失败'
  }
}
</script>

<style scoped>
.login-container {
  width: 400px;
  margin: 100px auto;
  padding: 30px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.1);
}
.login-form input {
  width: 100%;
  padding: 10px;
  margin: 8px 0;
  box-sizing: border-box;
}
.login-form button {
  width: 100%;
  padding: 10px;
  margin-top: 10px;
  background: #409eff;
  color: white;
  border: none;
  cursor: pointer;
}
.error {
  color: red;
}
</style>
