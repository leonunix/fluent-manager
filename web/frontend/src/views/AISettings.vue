<template>
  <div class="fm-ai-page">
    <div class="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-4">
      <div>
        <h4 class="mb-1">{{ t('ai_settings_page.title') }}</h4>
        <div class="text-muted">{{ t('ai_settings_page.intro_body') }}</div>
      </div>
      <button class="btn btn-primary" :disabled="!canUpdate || saving" @click="saveSettings">
        <i class="bi bi-save me-1"></i>{{ saving ? '...' : t('ai_settings_page.save_and_use') }}
      </button>
    </div>

    <div class="fm-ai-hero mb-4">
      <div>
        <div class="fm-ai-hero__eyebrow">{{ t('ai_settings_page.intro_title') }}</div>
        <div class="small text-muted">{{ t('ai_settings_page.enabled_hint') }}</div>
      </div>
      <div class="fm-ai-hero__stats">
        <div class="fm-ai-stat">
          <span class="fm-ai-stat__value">{{ form.accounts.length }}</span>
          <span class="fm-ai-stat__label">{{ t('ai_settings_page.account_count') }}</span>
        </div>
        <div class="fm-ai-stat">
          <span class="fm-ai-stat__value">{{ activeAccount?.name || t('ai_settings_page.no_active_account') }}</span>
          <span class="fm-ai-stat__label">{{ t('ai_settings_page.active_account') }}</span>
        </div>
      </div>
    </div>

    <div class="card border-0 shadow-sm mb-4">
      <div class="card-body">
        <div class="d-flex flex-wrap justify-content-between align-items-start gap-3 mb-3">
          <div>
            <div class="fw-semibold">{{ t('ai_settings_page.settings_title') }}</div>
            <div class="small text-muted">{{ t('ai_settings_page.save_hint') }}</div>
          </div>
          <div class="d-flex flex-wrap gap-2">
            <button type="button" class="btn btn-sm btn-outline-primary" :disabled="!canUpdate" @click="addAccount('openai')">{{ t('ai_settings_page.add_openai') }}</button>
            <button type="button" class="btn btn-sm btn-outline-primary" :disabled="!canUpdate" @click="addAccount('claude')">{{ t('ai_settings_page.add_claude') }}</button>
            <button type="button" class="btn btn-sm btn-outline-primary" :disabled="!canUpdate" @click="addAccount('gemini')">{{ t('ai_settings_page.add_gemini') }}</button>
            <button type="button" class="btn btn-sm btn-outline-primary" :disabled="!canUpdate" @click="addAccount('deepseek')">{{ t('ai_settings_page.add_deepseek') }}</button>
          </div>
        </div>

        <div class="row g-3 mb-3">
          <div class="col-md-3">
            <div class="form-check form-switch mt-md-4">
              <input v-model="form.enabled" class="form-check-input" type="checkbox" :disabled="!canUpdate">
              <label class="form-check-label">{{ t('auth_settings.enabled') }}</label>
            </div>
          </div>
          <div class="col-md-3">
            <label class="form-label">{{ t('auth_settings.request_timeout') }}</label>
            <input v-model.number="form.request_timeout_seconds" type="number" min="1" class="form-control" :disabled="!canUpdate">
          </div>
          <div class="col-md-6">
            <label class="form-label">{{ t('ai_settings_page.active_account') }}</label>
            <select v-model="form.active_account_id" class="form-select" :disabled="!canUpdate">
              <option value="">{{ t('ai_settings_page.no_active_account') }}</option>
              <option v-for="account in form.accounts" :key="account.id" :value="account.id">
                {{ account.name }} / {{ providerLabel(account.provider) }}
              </option>
            </select>
          </div>
          <div class="col-md-12">
            <label class="form-label">{{ t('auth_settings.system_prompt') }}</label>
            <textarea v-model="form.system_prompt" class="form-control" rows="3" :disabled="!canUpdate" :placeholder="t('auth_settings.system_prompt_placeholder')"></textarea>
            <div class="form-text">{{ t('ai_settings_page.system_prompt_help') }}</div>
          </div>
        </div>

        <div v-if="!form.accounts.length" class="fm-ai-empty">
          <i class="bi bi-stars"></i>
          <span>{{ t('ai_settings_page.no_accounts') }}</span>
        </div>

        <div v-else class="fm-ai-account-grid">
          <div v-for="account in form.accounts" :key="account.id" class="fm-ai-account-card" :class="{ 'is-active': form.active_account_id === account.id }">
            <div class="fm-ai-account-card__header">
              <div>
                <div class="fw-semibold">{{ account.name || providerLabel(account.provider) }}</div>
                <div class="small text-muted">{{ providerLabel(account.provider) }}</div>
              </div>
              <div class="d-flex align-items-center gap-2 flex-wrap justify-content-end">
                <span v-if="form.active_account_id === account.id" class="badge text-bg-primary">{{ t('ai_settings_page.active_badge') }}</span>
                <button type="button" class="btn btn-sm btn-outline-primary" :disabled="!canUpdate" @click="setActiveAccount(account.id)">{{ t('ai_settings_page.set_active') }}</button>
                <button type="button" class="btn btn-sm btn-outline-danger" :disabled="!canUpdate" @click="removeAccount(account.id)">{{ t('ai_settings_page.remove_account') }}</button>
              </div>
            </div>

            <div class="row g-3">
              <div class="col-md-6">
                <label class="form-label">{{ t('ai_settings_page.account_name') }}</label>
                <input v-model="account.name" type="text" class="form-control" :disabled="!canUpdate">
              </div>
              <div class="col-md-6">
                <label class="form-label">{{ t('ai_settings_page.provider') }}</label>
                <select v-model="account.provider" class="form-select" :disabled="!canUpdate">
                  <option value="openai">OpenAI</option>
                  <option value="claude">Claude</option>
                  <option value="gemini">Gemini</option>
                  <option value="deepseek">DeepSeek</option>
                </select>
              </div>
              <div class="col-md-12">
                <label class="form-label">{{ t('ai_settings_page.account_description') }}</label>
                <input v-model="account.description" type="text" class="form-control" :disabled="!canUpdate">
              </div>
              <div class="col-md-12">
                <div class="form-check form-switch">
                  <input v-model="account.enabled" class="form-check-input" type="checkbox" :disabled="!canUpdate">
                  <label class="form-check-label">{{ t('auth_settings.provider_enabled') }}</label>
                </div>
              </div>
              <div class="col-md-12">
                <label class="form-label">{{ t('auth_settings.api_key') }}</label>
                <input v-model="account.api_key" type="password" class="form-control" :disabled="!canUpdate" :placeholder="t('auth_settings.api_key_placeholder')">
              </div>
              <div class="col-md-12">
                <label class="form-label">{{ t('auth_settings.base_url') }}</label>
                <input v-model="account.base_url" type="text" class="form-control" :disabled="!canUpdate" :placeholder="providerPlaceholder(account.provider, 'base_url')">
              </div>
              <div class="col-md-12">
                <label class="form-label">{{ t('auth_settings.model') }}</label>
                <input v-model="account.model" type="text" class="form-control" :disabled="!canUpdate" :placeholder="providerPlaceholder(account.provider, 'model')">
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { getAISettings, updateAISettings } from '../api'
import { useI18n } from '../i18n'
import { useAuthStore } from '../store/auth'

const { t } = useI18n()
const auth = useAuthStore()

const providerMeta = {
  openai: { label: 'OpenAI', base_url: 'https://api.openai.com/v1', model: 'gpt-...' },
  claude: { label: 'Claude', base_url: 'https://api.anthropic.com', model: 'claude-...' },
  gemini: { label: 'Gemini', base_url: 'https://generativelanguage.googleapis.com', model: 'gemini-...' },
  deepseek: { label: 'DeepSeek', base_url: 'https://api.deepseek.com/v1', model: 'deepseek-...' },
}

function createForm() {
  return {
    enabled: false,
    active_provider: '',
    active_account_id: '',
    request_timeout_seconds: 60,
    system_prompt: '',
    accounts: [],
  }
}

function createAccount(provider = 'openai') {
  return {
    id: `acc-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    name: `${providerMeta[provider].label} Account`,
    provider,
    description: '',
    enabled: true,
    api_key: '',
    base_url: '',
    model: '',
  }
}

const form = reactive(createForm())
const saving = ref(false)

const canUpdate = computed(() => auth.hasPermission('ai_settings', 'update'))
const activeAccount = computed(() => form.accounts.find(account => account.id === form.active_account_id) || null)

function providerLabel(provider) {
  return providerMeta[provider]?.label || provider
}

function providerPlaceholder(provider, field) {
  return providerMeta[provider]?.[field] || ''
}

function ensureSettings() {
  if (!form.request_timeout_seconds || form.request_timeout_seconds < 1) form.request_timeout_seconds = 60
  if (!Array.isArray(form.accounts)) form.accounts = []
  if (!form.active_account_id && form.accounts.length) form.active_account_id = form.accounts[0].id
  const active = form.accounts.find(account => account.id === form.active_account_id)
  form.active_provider = active?.provider || ''
}

function clonePlain(value) {
  return JSON.parse(JSON.stringify(value))
}

async function loadSettings() {
  const { data } = await getAISettings()
  Object.assign(form, createForm(), data)
  form.accounts = (data.accounts || []).map(account => ({
    id: account.id,
    name: account.name || `${providerLabel(account.provider)} Account`,
    provider: account.provider || 'openai',
    description: account.description || '',
    enabled: account.enabled !== false,
    api_key: account.api_key || '',
    base_url: account.base_url || '',
    model: account.model || '',
  }))
  ensureSettings()
}

function addAccount(provider) {
  form.accounts.push(createAccount(provider))
  if (!form.active_account_id) {
    form.active_account_id = form.accounts[form.accounts.length - 1].id
  }
  ensureSettings()
}

function removeAccount(accountID) {
  form.accounts = form.accounts.filter(account => account.id !== accountID)
  if (form.active_account_id === accountID) {
    form.active_account_id = form.accounts[0]?.id || ''
  }
  ensureSettings()
}

function setActiveAccount(accountID) {
  form.active_account_id = accountID
  ensureSettings()
}

async function saveSettings() {
  if (!canUpdate.value) return
  saving.value = true
  try {
    ensureSettings()
    await updateAISettings(clonePlain(form))
    await loadSettings()
    alert(t('auth_settings.save_success'))
  } finally {
    saving.value = false
  }
}

onMounted(loadSettings)
</script>

<style scoped>
.fm-ai-page { max-width: 1400px; }
.fm-ai-hero {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  padding: 1.2rem 1.25rem;
  border-radius: 20px;
  border: 1px solid #dbe6f3;
  background: linear-gradient(135deg, rgba(219, 234, 254, 0.88) 0%, rgba(248, 250, 252, 0.96) 56%, #ffffff 100%);
}
.fm-ai-hero__eyebrow {
  font-size: 0.78rem;
  font-weight: 700;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: #2563eb;
  margin-bottom: 0.35rem;
}
.fm-ai-hero__stats {
  display: flex;
  gap: 0.8rem;
  flex-wrap: wrap;
}
.fm-ai-stat {
  min-width: 180px;
  padding: 0.9rem 1rem;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(148, 163, 184, 0.18);
}
.fm-ai-stat__value {
  display: block;
  font-weight: 700;
  color: #0f172a;
}
.fm-ai-stat__label {
  display: block;
  margin-top: 0.2rem;
  font-size: 0.75rem;
  color: #64748b;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.fm-ai-empty {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.95rem 1rem;
  border-radius: 16px;
  border: 1px dashed rgba(148, 163, 184, 0.4);
  color: #64748b;
}
.fm-ai-account-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 1rem;
}
.fm-ai-account-card {
  padding: 1rem;
  border-radius: 18px;
  border: 1px solid #dbe6f3;
  background: #ffffff;
  box-shadow: 0 10px 28px rgba(15, 23, 42, 0.04);
}
.fm-ai-account-card.is-active {
  border-color: #2563eb;
  background: linear-gradient(180deg, rgba(219, 234, 254, 0.5) 0%, #ffffff 100%);
  box-shadow: 0 16px 36px rgba(37, 99, 235, 0.12);
}
.fm-ai-account-card__header {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: flex-start;
  margin-bottom: 1rem;
}
@media (max-width: 991.98px) {
  .fm-ai-hero { flex-direction: column; }
  .fm-ai-account-grid { grid-template-columns: 1fr; }
}
</style>
