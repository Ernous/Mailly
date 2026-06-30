import { createRouter, createWebHashHistory } from 'vue-router'
import LoginPage from './components/LoginPage.vue'
import MailView from './components/MailView.vue'

const routes = [
  { path: '/', redirect: '/mail' },
  { path: '/login', name: 'login', component: LoginPage },
  { path: '/mail', name: 'mail', component: MailView },
  { path: '/mail/:accountId', name: 'mail-account', component: MailView },
  { path: '/mail/:accountId/:folder', name: 'mail-folder', component: MailView },
]

const router = createRouter({
  history: createWebHashHistory(),
  routes,
})

export default router
