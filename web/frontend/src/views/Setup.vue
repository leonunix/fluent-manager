<template>
  <div class="fm-login-page">
    <div class="fm-login-left">
      <div class="fm-login-brand">
        <div class="fm-login-logo">
          <i class="bi bi-diagram-3-fill"></i>
        </div>
        <h1>Fluent Manager</h1>
        <p>Enterprise Log Agent Management Platform</p>
      </div>
      <div class="fm-login-features">
        <div class="fm-feature-item">
          <i class="bi bi-rocket-takeoff"></i>
          <span>{{ t('setup.welcome_desc') }}</span>
        </div>
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
      </div>
    </div>
    <div class="fm-login-right">
      <div class="fm-login-card" style="max-width:460px">
        <h3>{{ t('setup.title') }}</h3>
        <p class="text-muted mb-3">{{ t('setup.subtitle') }}</p>

        <!-- Step indicator -->
        <div class="fm-steps mb-4">
          <div class="fm-step" :class="{ active: step === 1, done: step > 1 }">
            <span class="fm-step-num">{{ step > 1 ? '✓' : '1' }}</span>
            <span class="fm-step-label">{{ t('setup.step_database') }}</span>
          </div>
          <div class="fm-step-line" :class="{ done: step > 1 }"></div>
          <div class="fm-step" :class="{ active: step === 2 }">
            <span class="fm-step-num">2</span>
            <span class="fm-step-label">{{ t('setup.step_admin') }}</span>
          </div>
        </div>

        <!-- Restarting overlay -->
        <div v-if="restarting" class="text-center py-5">
          <div class="spinner-border text-primary mb-3" style="width:3rem;height:3rem"></div>
          <p class="text-muted">{{ t('setup.restarting') }}</p>
        </div>

        <template v-else>
          <div v-if="error" class="alert alert-danger py-2">{{ error }}</div>
          <div v-if="success" class="alert alert-success py-2">{{ t('setup.success') }}</div>

          <!-- Step 1: Database -->
          <form v-if="step === 1" @submit.prevent="goStep2">
            <div class="mb-3">
              <label class="form-label">{{ t('setup.db_driver') }}</label>
              <select v-model="dbForm.driver" class="form-select">
                <option value="sqlite">{{ t('setup.db_sqlite') }}</option>
                <option value="mysql">{{ t('setup.db_mysql') }}</option>
                <option value="postgres">{{ t('setup.db_postgres') }}</option>
              </select>
            </div>

            <!-- SQLite -->
            <div v-if="dbForm.driver === 'sqlite'" class="mb-3">
              <label class="form-label">{{ t('setup.db_path') }}</label>
              <input v-model="dbForm.path" type="text" class="form-control" :placeholder="t('setup.db_path_ph')">
            </div>

            <!-- MySQL / PostgreSQL -->
            <template v-if="dbForm.driver === 'mysql' || dbForm.driver === 'postgres'">
              <div class="row mb-3">
                <div class="col-8">
                  <label class="form-label">{{ t('setup.db_host') }}</label>
                  <input v-model="dbForm.host" type="text" class="form-control" :placeholder="t('setup.db_host_ph')" required>
                </div>
                <div class="col-4">
                  <label class="form-label">{{ t('setup.db_port') }}</label>
                  <input v-model.number="dbForm.port" type="number" class="form-control" required>
                </div>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('setup.db_user') }}</label>
                <input v-model="dbForm.user" type="text" class="form-control" :placeholder="t('setup.db_user_ph')" required>
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('setup.db_password') }}</label>
                <input v-model="dbForm.password" type="password" class="form-control" :placeholder="t('setup.db_password_ph')">
              </div>
              <div class="mb-3">
                <label class="form-label">{{ t('setup.db_name') }}</label>
                <input v-model="dbForm.dbName" type="text" class="form-control" :placeholder="t('setup.db_name_ph')" required>
              </div>
            </template>

            <!-- Test connection button -->
            <div v-if="dbForm.driver !== 'sqlite'" class="mb-3">
              <button type="button" class="btn btn-outline-secondary btn-sm" :disabled="testingDB" @click="handleTestDB">
                <span v-if="testingDB" class="spinner-border spinner-border-sm me-1"></span>
                {{ testingDB ? t('setup.testing') : t('setup.test_connection') }}
              </button>
              <span v-if="testResult === true" class="ms-2 text-success"><i class="bi bi-check-circle"></i> {{ t('setup.test_success') }}</span>
              <span v-if="testResult === false" class="ms-2 text-danger"><i class="bi bi-x-circle"></i> {{ testError }}</span>
            </div>

            <button type="submit" class="btn btn-primary w-100 py-2">
              {{ t('setup.next') }} <i class="bi bi-arrow-right"></i>
            </button>
          </form>

          <!-- Step 2: Admin Account -->
          <form v-if="step === 2" @submit.prevent="handleSetup">
            <div class="mb-3">
              <label class="form-label">{{ t('setup.username') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-person"></i></span>
                <input v-model="adminForm.username" type="text" class="form-control" :placeholder="t('setup.username_ph')" required autofocus>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('setup.password') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-lock"></i></span>
                <input v-model="adminForm.password" type="password" class="form-control" :placeholder="t('setup.password_ph')" required>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('setup.confirm_password') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-lock-fill"></i></span>
                <input v-model="adminForm.confirmPassword" type="password" class="form-control" :placeholder="t('setup.confirm_password_ph')" required>
              </div>
            </div>
            <div class="mb-3">
              <label class="form-label">{{ t('setup.email') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-envelope"></i></span>
                <input v-model="adminForm.email" type="email" class="form-control" :placeholder="t('setup.email_ph')">
              </div>
            </div>
            <div class="mb-4">
              <label class="form-label">{{ t('setup.display_name') }}</label>
              <div class="input-group">
                <span class="input-group-text"><i class="bi bi-tag"></i></span>
                <input v-model="adminForm.displayName" type="text" class="form-control" :placeholder="t('setup.display_name_ph')">
              </div>
            </div>
            <div class="d-flex gap-2">
              <button type="button" class="btn btn-outline-secondary py-2" @click="step = 1">
                <i class="bi bi-arrow-left"></i> {{ t('setup.back') }}
              </button>
              <button type="submit" class="btn btn-primary flex-fill py-2" :disabled="loading">
                <span v-if="loading" class="spinner-border spinner-border-sm me-1"></span>
                {{ t('setup.btn') }}
              </button>
            </div>
          </form>
        </template>

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
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'
import { useI18n } from '../i18n'
import { testDBConnection, initializeSystem, getSetupStatus } from '../api/setup'
import { resetSetupCache } from '../router'

const router = useRouter()
const auth = useAuthStore()
const { t, locale: currentLocale, setLocale, availableLocales, localeNames } = useI18n()

const step = ref(1)
const loading = ref(false)
const restarting = ref(false)
const error = ref('')
const success = ref(false)
const locale = ref(currentLocale.value)

// DB test state
const testingDB = ref(false)
const testResult = ref(null) // null=untested, true=ok, false=failed
const testError = ref('')

const dbForm = reactive({
  driver: 'sqlite',
  path: '',
  host: 'localhost',
  port: 3306,
  user: '',
  password: '',
  dbName: 'fluent_manager',
})

const adminForm = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  displayName: '',
})

async function handleTestDB() {
  testingDB.value = true
  testResult.value = null
  testError.value = ''
  try {
    await testDBConnection({
      driver: dbForm.driver,
      host: dbForm.host,
      port: dbForm.port,
      user: dbForm.user,
      password: dbForm.password,
      db_name: dbForm.dbName,
      path: dbForm.path,
    })
    testResult.value = true
  } catch (e) {
    testResult.value = false
    testError.value = e.response?.data?.error || t('setup.test_failed')
  } finally {
    testingDB.value = false
  }
}

function goStep2() {
  error.value = ''
  // Update default port when switching drivers
  if (dbForm.driver === 'postgres' && dbForm.port === 3306) {
    dbForm.port = 5432
  }
  if (dbForm.driver === 'mysql' && dbForm.port === 5432) {
    dbForm.port = 3306
  }
  step.value = 2
}

async function handleSetup() {
  error.value = ''

  if (adminForm.password.length < 8) {
    error.value = t('setup.password_too_short')
    return
  }
  if (adminForm.password !== adminForm.confirmPassword) {
    error.value = t('setup.password_mismatch')
    return
  }

  loading.value = true
  try {
    const payload = {
      username: adminForm.username,
      password: adminForm.password,
      email: adminForm.email || undefined,
      display_name: adminForm.displayName || undefined,
    }

    // Include DB config if not default SQLite
    if (dbForm.driver !== 'sqlite' || (dbForm.path && dbForm.path !== 'fluent_manager.db')) {
      payload.db_driver = dbForm.driver
      if (dbForm.driver === 'sqlite') {
        payload.db_path = dbForm.path
      } else {
        payload.db_host = dbForm.host
        payload.db_port = dbForm.port
        payload.db_user = dbForm.user
        payload.db_password = dbForm.password
        payload.db_name = dbForm.dbName
      }
    }

    const res = await initializeSystem(payload)
    const { token, user, restart } = res.data

    // Store auth data and mark setup as complete
    localStorage.setItem('token', token)
    localStorage.setItem('user', JSON.stringify(user))
    auth.$patch({ token, user })
    resetSetupCache()

    if (restart) {
      // Server is restarting — poll until it's back
      success.value = true
      restarting.value = true
      await waitForRestart()
      router.push('/')
    } else {
      // No restart needed — go straight to dashboard
      success.value = true
      setTimeout(() => router.push('/'), 800)
    }
  } catch (e) {
    error.value = e.response?.data?.error || 'Setup failed'
  } finally {
    loading.value = false
  }
}

async function waitForRestart() {
  const maxAttempts = 30
  for (let i = 0; i < maxAttempts; i++) {
    await sleep(2000)
    try {
      const res = await getSetupStatus()
      if (res.data.initialized) {
        return
      }
    } catch {
      // Server still restarting, keep polling
    }
  }
}

function sleep(ms) {
  return new Promise(resolve => setTimeout(resolve, ms))
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
  width: 56px;
  height: 56px;
  background: linear-gradient(135deg, #3b82f6, #2563eb);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 1.5rem;
  margin-bottom: 24px;
  box-shadow: 0 4px 16px rgba(59, 130, 246, 0.4);
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
  max-width: 460px;
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
/* Step indicator */
.fm-steps {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0;
}
.fm-step {
  display: flex;
  align-items: center;
  gap: 6px;
  opacity: 0.5;
}
.fm-step.active, .fm-step.done {
  opacity: 1;
}
.fm-step-num {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #e2e8f0;
  color: #64748b;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 600;
}
.fm-step.active .fm-step-num {
  background: #3b82f6;
  color: #fff;
}
.fm-step.done .fm-step-num {
  background: #22c55e;
  color: #fff;
}
.fm-step-label {
  font-size: 0.85rem;
  color: #64748b;
  font-weight: 500;
}
.fm-step.active .fm-step-label {
  color: #1e293b;
}
.fm-step-line {
  width: 40px;
  height: 2px;
  background: #e2e8f0;
  margin: 0 8px;
}
.fm-step-line.done {
  background: #22c55e;
}
@media (max-width: 768px) {
  .fm-login-left { display: none; }
  .fm-login-right { padding: 20px; }
}
</style>
