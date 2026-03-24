import type { OutputTarget } from '../types/fluent'

interface OutputSettings {
  index?: string
  match?: string
  tenant_id?: string
  labels?: string
  topics?: string
  format?: string
  uri?: string
  bucket?: string
  path?: string
  compression?: string
  plugin?: string
  tls?: boolean | string
  scheme?: string
  http_user?: string
  http_password?: string
  user?: string
  password?: string
  header_authorization?: string
  sasl_username?: string
  sasl_password?: string
  [key: string]: unknown
}

export interface OutputSummary {
  settings: OutputSettings
  chips: string[]
  endpoint: string
  primary: string
  secondary: string[]
}

function parseSettings(raw: string | undefined | null): OutputSettings {
  if (!raw) return {}
  try {
    const parsed = JSON.parse(raw)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

function endpointLooksTLS(endpoint: string | undefined | null): boolean {
  return String(endpoint || '').toLowerCase().startsWith('https://')
}

function hasAuth(settings: OutputSettings): boolean {
  return Boolean(
    settings.http_user ||
    settings.http_password ||
    settings.user ||
    settings.password ||
    settings.header_authorization ||
    settings.sasl_username ||
    settings.sasl_password
  )
}

function hasTLS(settings: OutputSettings, endpoint: string | undefined | null): boolean {
  return settings.tls === true || settings.tls === 'On' || settings.scheme === 'https' || endpointLooksTLS(endpoint)
}

function pushChip(chips: string[], value: string | undefined | null): void {
  if (!value) return
  const normalized = String(value).trim()
  if (!normalized) return
  if (!chips.includes(normalized)) {
    chips.push(normalized)
  }
}

export function summarizeOutputTarget(target: OutputTarget | null | undefined): OutputSummary {
  const settings = parseSettings(target?.settings)
  const chips: string[] = []

  if (target?.target_type === 'opensearch') {
    pushChip(chips, settings.index ? `index:${settings.index}` : '')
    pushChip(chips, settings.match ? `match:${settings.match}` : '')
  } else if (target?.target_type === 'loki') {
    pushChip(chips, settings.tenant_id ? `tenant:${settings.tenant_id}` : '')
    pushChip(chips, settings.labels ? `labels:${settings.labels}` : '')
  } else if (target?.target_type === 'kafka') {
    pushChip(chips, settings.topics ? `topic:${settings.topics}` : '')
    pushChip(chips, settings.format ? `format:${settings.format}` : '')
  } else if (target?.target_type === 'http') {
    pushChip(chips, settings.uri ? `uri:${settings.uri}` : '')
    pushChip(chips, settings.format ? `format:${settings.format}` : '')
  } else if (target?.target_type === 's3') {
    pushChip(chips, settings.bucket ? `bucket:${settings.bucket}` : '')
    pushChip(chips, settings.path ? `path:${settings.path}` : '')
    pushChip(chips, settings.compression ? `compression:${settings.compression}` : '')
  } else if (target?.target_type === 'stdout') {
    pushChip(chips, settings.format ? `format:${settings.format}` : '')
  } else {
    pushChip(chips, settings.plugin ? `plugin:${settings.plugin}` : '')
    pushChip(chips, settings.match ? `match:${settings.match}` : '')
  }

  if (hasTLS(settings, target?.endpoint)) {
    pushChip(chips, 'TLS')
  }
  if (hasAuth(settings)) {
    pushChip(chips, 'AUTH')
  }

  return {
    settings,
    chips,
    endpoint: target?.endpoint || '-',
    primary: chips[0] || '',
    secondary: chips.slice(1, 4),
  }
}

export function parseOutputTargetSettings(raw: string | undefined | null): OutputSettings {
  return parseSettings(raw)
}
