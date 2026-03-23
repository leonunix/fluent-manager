import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import * as bootstrap from 'bootstrap'

import 'bootstrap/dist/css/bootstrap.min.css'
import 'bootstrap-icons/font/bootstrap-icons.css'
import './assets/theme.css'

declare global {
  interface Window {
    bootstrap: typeof bootstrap
  }
}

// Existing views access Bootstrap modal constructors via window.bootstrap.
if (typeof window !== 'undefined') {
  window.bootstrap = bootstrap
}

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')
