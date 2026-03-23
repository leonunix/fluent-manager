<template>
  <div class="fm-login-page">
    <div class="fm-login-left">
      <div class="fm-login-brand">
        <div class="fm-login-logo">
          <img src="/brand/logo-flow-mesh.svg" alt="Fluent Manager logo">
        </div>
        <h1>Fluent Manager</h1>
        <p>Enterprise Log Agent Management Platform</p>
      </div>
      <div class="fm-login-features">
        <div class="fm-feature-item">
          <i class="bi bi-building"></i>
          <span>Multi-datacenter topology management</span>
        </div>
        <div class="fm-feature-item">
          <i class="bi bi-gear-wide-connected"></i>
          <span>Centralized config deployment &amp; versioning</span>
        </div>
        <div class="fm-feature-item">
          <i class="bi bi-shield-check"></i>
          <span>Resource-scoped access control</span>
        </div>
        <div class="fm-feature-item">
          <i class="bi bi-activity"></i>
          <span>Real-time monitoring &amp; remote control</span>
        </div>
      </div>
    </div>
    <div class="fm-login-right">
      <div class="fm-login-card">
        <h3>{{ t('login.title') }}</h3>
        <p class="text-muted mb-4">{{ t('login.subtitle') }}</p>
        <div v-if="error" class="alert alert-danger py-2">{{ error }}</div>
        <form @submit.prevent="handleLogin">
          <div class="mb-3">
            <label class="form-label">{{ t('login.username') }}</label>
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-person"></i></span>
              <input v-model="form.username" type="text" class="form-control" :placeholder="t('login.username_ph')" required autofocus>
            </div>
          </div>
          <div class="mb-3">
            <label class="form-label">{{ t('login.password') }}</label>
            <div class="input-group">
              <span class="input-group-text"><i class="bi bi-lock"></i></span>
              <input v-model="form.password" type="password" class="form-control" :placeholder="t('login.password_ph')" required>
            </div>
          </div>
          <div v-if="passwordMethods.length > 1" class="mb-4">
            <label class="form-label">{{ t('login.auth_method') }}</label>
            <select v-model="form.authSource" class="form-select">
              <option v-for="m in passwordMethods" :key="m" :value="m">{{ methodLabel(m) }}</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary w-100 py-2" :disabled="loading">
            <span v-if="loading" class="spinner-border spinner-border-sm me-1"></span>
            {{ t('login.btn') }}
          </button>
        </form>
        <div v-if="authMethods.includes('saml')" class="mt-3">
          <div class="text-center text-muted mb-2" style="font-size:0.85rem">{{ t('login.or_divider') || 'or' }}</div>
          <a href="/saml/login" class="btn btn-outline-secondary w-100 py-2">
            <i class="bi bi-shield-lock me-1"></i> {{ t('login.saml_btn') || 'Login with SSO (SAML)' }}
          </a>
        </div>
        <div class="fm-login-footer">
          <select v-model="locale" class="form-select form-select-sm" style="width:auto" @change="setLocale(locale)">
            <option v-for="l in availableLocales" :key="l" :value="l">{{ localeNames[l] }}</option>
          </select>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { useI18n } from '../i18n'
import { getAuthMethods } from '../api/auth'

const router = useRouter()
const auth = useAuthStore()
const { t, locale: currentLocale, setLocale, availableLocales, localeNames } = useI18n()
const loading = ref(false)
const error = ref('')
const locale = ref(currentLocale.value)
const authMethods = ref(['local'])

const form = reactive({ username: '', password: '', authSource: 'local' })

// Password-based methods only (SAML uses redirect, shown as a separate button)
const passwordMethods = computed(() => authMethods.value.filter(m => m !== 'saml'))

const methodLabel = (m) => {
  if (m === 'local') return t('login.auth_local')
  if (m === 'ldap') return 'LDAP'
  if (m === 'saml') return 'SAML'
  return m
}

onMounted(async () => {
  try {
    const res = await getAuthMethods()
    authMethods.value = res.data.methods || ['local']
  } catch {
    authMethods.value = ['local']
  }
})

async function handleLogin() {
  loading.value = true
  error.value = ''
  localStorage.setItem('fm_locale', locale.value)
  try {
    await auth.login(form.username, form.password, form.authSource)
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.fm-login-page {
  display: flex;
  min-height: 100vh;
}
.fm-login-left {
  flex: 0 0 45%;
  background: linear-gradient(135deg, #1e3a5f 0%, #0f172a 50%, #1e293b 100%);
  display: flex;
  flex-direction: column;
  justify-content: center;
  padding: 60px;
  position: relative;
  overflow: hidden;
}
.fm-login-left::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -30%;
  width: 600px;
  height: 600px;
  background: radial-gradient(circle, rgba(59, 130, 246, 0.12) 0%, transparent 70%);
  border-radius: 50%;
}
.fm-login-left::after {
  content: '';
  position: absolute;
  bottom: -30%;
  left: -20%;
  width: 400px;
  height: 400px;
  background: radial-gradient(circle, rgba(99, 102, 241, 0.1) 0%, transparent 70%);
  border-radius: 50%;
}
.fm-login-brand {
  position: relative;
  z-index: 1;
  margin-bottom: 48px;
}
.fm-login-logo {
  width: 64px;
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}
.fm-login-logo img {
  width: 100%;
  height: 100%;
  display: block;
  filter: drop-shadow(0 10px 24px rgba(37, 99, 235, 0.28));
}
.fm-login-brand h1 {
  font-size: 2rem;
  font-weight: 700;
  color: #f1f5f9;
  margin-bottom: 8px;
  letter-spacing: -0.02em;
}
.fm-login-brand p {
  font-size: 0.95rem;
  color: #64748b;
}
.fm-login-features {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.fm-feature-item {
  display: flex;
  align-items: center;
  gap: 14px;
  color: #94a3b8;
  font-size: 0.9rem;
}
.fm-feature-item i {
  font-size: 1.1rem;
  color: #60a5fa;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: rgba(59, 130, 246, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.fm-login-right {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f8fafc;
  padding: 40px;
}
.fm-login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 16px;
  padding: 40px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.06);
  border: 1px solid #e2e8f0;
}
.fm-login-card h3 {
  font-weight: 700;
  color: #1e293b;
  margin-bottom: 4px;
}
.fm-login-card .input-group-text {
  background: #f8fafc;
  border-color: #d1d5db;
  color: #94a3b8;
}
.fm-login-footer {
  margin-top: 24px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
  display: flex;
  justify-content: center;
}
@media (max-width: 768px) {
  .fm-login-left { display: none; }
  .fm-login-right { padding: 20px; }
}
</style>
