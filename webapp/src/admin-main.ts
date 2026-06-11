import { createApp } from 'vue'
import AdminApp from './AdminApp.vue'
import router from './router/admin'
import i18n from './i18n'
import './styles/theme.css'
import './style.css'
import './styles/markdown-content.css'

const app = createApp(AdminApp)

app.config.errorHandler = (err, _instance, info) => {
  console.error('[Vue Error]', err, info)
}

app.use(router)
app.use(i18n)
app.mount('#app')
