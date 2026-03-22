<template>
  <div class="min-vh-100 d-flex align-items-center justify-content-center bg-light">
    <div class="card shadow" style="width: 400px;">
      <div class="card-body p-4">
        <h3 class="text-center mb-4">
          <i class="bi bi-diagram-3-fill text-primary"></i> Fluent Manager
        </h3>
        <div v-if="error" class="alert alert-danger">{{ error }}</div>
        <form @submit.prevent="handleLogin">
          <div class="mb-3">
            <label class="form-label">用户名</label>
            <input v-model="form.username" type="text" class="form-control" required autofocus>
          </div>
          <div class="mb-3">
            <label class="form-label">密码</label>
            <input v-model="form.password" type="password" class="form-control" required>
          </div>
          <div class="mb-3">
            <label class="form-label">认证方式</label>
            <select v-model="form.authSource" class="form-select">
              <option value="local">本地认证</option>
              <option value="ldap">LDAP</option>
            </select>
          </div>
          <button type="submit" class="btn btn-primary w-100" :disabled="loading">
            <span v-if="loading" class="spinner-border spinner-border-sm me-1"></span>
            登录
          </button>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../store/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)
const error = ref('')

const form = reactive({
  username: '',
  password: '',
  authSource: 'local',
})

async function handleLogin() {
  loading.value = true
  error.value = ''
  try {
    await auth.login(form.username, form.password, form.authSource)
    router.push('/')
  } catch (e) {
    error.value = e.response?.data?.error || '登录失败'
  } finally {
    loading.value = false
  }
}
</script>
