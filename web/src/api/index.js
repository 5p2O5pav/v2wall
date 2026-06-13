import axios from 'axios'
import router from '../router'

// 根据部署时的 admin_path 配置前缀，开发时通过 vite proxy 代理
const BASE_URL = '/admin-a3f2b1c5'  // 如果打包后使用 Nginx 反代，确保路径正确

const api = axios.create({
  baseURL: BASE_URL,
  timeout: 15000
})

// 请求拦截：附加 JWT
api.interceptors.request.use(config => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截：处理 401
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      localStorage.removeItem('token')
      router.push('/login')
    }
    return Promise.reject(error)
  }
)

// ---------- 接口定义 ----------
export function initAdmin(username, password, initToken) {
  return api.post('/api/v1/init', { username, password }, {
    headers: { Authorization: `Bearer ${initToken}` }
  })
}

export function login(username, password) {
  return api.post('/api/v1/login', { username, password })
}

export function getIPList(page = 1, size = 20) {
  return api.get('/api/v1/stats/ips', { params: { page, size } })
}

export function getIPLogs(ip, page = 1, size = 20) {
  return api.get('/api/v1/logs', { params: { ip, page, size } })
}

export function getIPInfo(ip) {
  return api.get('/api/v1/ipinfo', { params: { ip } })
}

export function getWhitelist() {
  return api.get('/api/v1/whitelist')
}

export function addWhitelist(cidr, remark) {
  return api.post('/api/v1/whitelist', { cidr, remark })
}

export function deleteWhitelist(cidr) {
  return api.delete(`/api/v1/whitelist/${encodeURIComponent(cidr)}`)
}

export function getConfig() {
  return api.get('/api/v1/config')
}

export function updateConfig(cfg) {
  return api.put('/api/v1/config', cfg)
}

export function getNodes() {
  return api.get('/api/v1/nodes')
}
