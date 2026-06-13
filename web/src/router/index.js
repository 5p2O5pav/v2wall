import { createRouter, createWebHashHistory } from 'vue-router'
import Layout from '../views/Layout.vue'
import IPList from '../views/IPList.vue'
import IPLogs from '../views/IPLogs.vue'
import Whitelist from '../views/Whitelist.vue'
import Config from '../views/Config.vue'
import Nodes from '../views/Nodes.vue'
import Login from '../views/Login.vue'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/',
    component: Layout,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        redirect: '/ips'
      },
      {
        path: 'ips',
        name: 'IPList',
        component: IPList
      },
      {
        path: 'logs',
        name: 'IPLogs',
        component: IPLogs
      },
      {
        path: 'whitelist',
        name: 'Whitelist',
        component: Whitelist
      },
      {
        path: 'config',
        name: 'Config',
        component: Config
      },
      {
        path: 'nodes',
        name: 'Nodes',
        component: Nodes
      }
    ]
  }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  if (to.matched.some(record => record.meta.requiresAuth)) {
    if (!token) {
      next({ name: 'Login', query: { redirect: to.fullPath } })
    } else {
      next()
    }
  } else {
    next()
  }
})

export default router
