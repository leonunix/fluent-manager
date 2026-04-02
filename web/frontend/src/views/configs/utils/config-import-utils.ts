// Import-config analysis utility functions.
// Pure functions – no reactive state. t-dependent functions accept `t` as last param.

import { normalizeName, sortObjectKeys, stringifySortedObject, uniqueSorted } from './config-text'
import { parseVariablesMap } from './config-variables'

export function importedModuleIdentity(name: string, moduleType: string, fluentType: string): string {
  return [normalizeName(name), String(moduleType || '').trim(), String(fluentType || '').trim()].join('::')
}

export function uniqueImportedModuleName(
  baseName: string,
  moduleType: string,
  fluentType: string,
  occupiedIdentities: Set<string>,
): string {
  const seed = normalizeName(baseName) ? String(baseName).trim() : 'imported-module'
  if (moduleType === 'parser') {
    return seed
  }
  const seedIdentity = importedModuleIdentity(seed, moduleType, fluentType)
  if (!occupiedIdentities.has(seedIdentity)) {
    occupiedIdentities.add(seedIdentity)
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (occupiedIdentities.has(importedModuleIdentity(candidate, moduleType, fluentType))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  occupiedIdentities.add(importedModuleIdentity(candidate, moduleType, fluentType))
  return candidate
}

export function findImportedModuleNameConflict(
  modules: any[],
  occupiedIdentities: Set<string>,
): { type: string; item: any } | null {
  const batchIdentities = new Set<string>()
  for (const item of modules || []) {
    if (item.module_type === 'output' || item.import_action === 'reuse_existing') {
      continue
    }
    const name = String(item.name || '').trim()
    if (!name) {
      continue
    }
    const identity = importedModuleIdentity(name, item.module_type, item.fluent_type)
    if (batchIdentities.has(identity)) {
      return {
        type: 'batch_duplicate',
        item,
      }
    }
    batchIdentities.add(identity)
    if (occupiedIdentities.has(identity) && item.module_type !== 'parser') {
      return {
        type: 'existing_duplicate',
        item,
      }
    }
  }
  return null
}

export function uniqueImportedOutputTargetName(baseName: string, occupiedNames: Set<string>): string {
  const seed = normalizeName(baseName) ? String(baseName).trim() : 'imported-destination'
  if (!occupiedNames.has(normalizeName(seed))) {
    occupiedNames.add(normalizeName(seed))
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (occupiedNames.has(normalizeName(candidate))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  occupiedNames.add(normalizeName(candidate))
  return candidate
}

export function normalizeImportedOutputModuleContent(content: string): string {
  return String(content || '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean)
    .join('\n')
}

export function createImportedOutputModuleSignature(module: any): string {
  return [
    String(module?.module_type || '').trim().toLowerCase(),
    String(module?.fluent_type || '').trim().toLowerCase(),
    normalizeImportedOutputModuleContent(module?.content),
  ].join('::')
}

export function pickImportedOutputTargetSettings(
  rawSettings: Record<string, any>,
  targetType: string,
): Record<string, any> {
  const settings = { ...(rawSettings || {}) }
  const keepByTargetType: Record<string, string[]> = {
    opensearch: [
      'host', 'port', 'http_user', 'http_passwd', 'cloud_id', 'path', 'uri',
      'tls', 'tls.verify', 'aws_auth', 'aws_region', 'aws_service_name',
    ],
    loki: ['host', 'port', 'uri', 'tenant_id', 'http_user', 'http_passwd', 'tls', 'tls.verify'],
    kafka: ['brokers', 'client_id', 'security.protocol', 'sasl.username', 'sasl.password', 'tls', 'tls.verify'],
    http: ['host', 'port', 'uri', 'http_user', 'http_passwd', 'proxy', 'tls', 'tls.verify'],
    s3: ['bucket', 'region', 'endpoint', 'host', 'port', 'uri', 'access_key_id', 'secret_access_key', 'role_arn', 'tls', 'tls.verify'],
    stdout: [],
  }
  const keepKeys = keepByTargetType[targetType] || []
  if (keepKeys.length) {
    return keepKeys.reduce((acc: Record<string, any>, key) => {
      if (settings[key] !== undefined && settings[key] !== null && String(settings[key]).trim() !== '') {
        acc[key] = settings[key]
      }
      return acc
    }, {})
  }

  const dropKeys = new Set([
    'match', 'logstash_format', 'logstash_prefix', 'logstash_dateformat', 'index', 'topic', 'topics',
    'labels', 'label_keys', 'remove_keys', 'generate_id', 'retry_limit', 'workers', 'compress',
    'replace_dots', 'suppress_type_name', 'trace_error', 'time_key', 'time_key_format',
  ])
  return Object.entries(settings).reduce((acc: Record<string, any>, [key, value]) => {
    if (dropKeys.has(key)) return acc
    if (value === undefined || value === null || String(value).trim() === '') return acc
    acc[key] = value
    return acc
  }, {})
}

export function createImportedOutputTargetSignature(target: any): string {
  const targetType = String(target?.target_type || '').trim().toLowerCase()
  const endpoint = String(target?.endpoint || '').trim().toLowerCase()
  const settings = typeof target?.settings === 'string'
    ? parseVariablesMap(target.settings)
    : (target?.settings || {})
  const filteredSettings = pickImportedOutputTargetSettings(settings, targetType)
  return [targetType, endpoint, stringifySortedObject(filteredSettings)].join('::')
}

export function uniqueImportedDestinationList(destinations: any[]): any[] {
  const seen = new Set<number>()
  const filtered: any[] = []
  for (const item of destinations || []) {
    const id = Number(item?.output_target_id || 0)
    if (!id || seen.has(id)) continue
    seen.add(id)
    filtered.push(item)
  }
  return filtered
}

export function findImportedOutputAdapterModule(
  targetType: string,
  modules: any[],
  fluentType: string,
): any | null {
  const normalizedType = String(targetType || '').trim().toLowerCase()
  if (!normalizedType) return null
  const matches = (modules || []).filter((item) =>
    item.module_type === 'output' &&
    item.preset_kind === 'output' &&
    String(item.preset_key || '').trim().toLowerCase() === normalizedType &&
    !!item.is_builtin
  )
  return matches.find((item) => item.fluent_type === fluentType) ||
    matches.find((item) => item.fluent_type === 'shared') ||
    matches[0] ||
    null
}

export function normalizeImportedOutputRenderVariables(
  targetType: string,
  fluentType: string,
  rawVariables: Record<string, any>,
): Record<string, any> {
  const normalizedType = String(targetType || '').trim().toLowerCase()
  const normalizedRuntime = String(fluentType || '').trim().toLowerCase()
  const variables = { ...(rawVariables || {}) }

  if (variables.http_passwd !== undefined && variables.http_password === undefined) {
    variables.http_password = variables.http_passwd
  }
  if (variables.http_password !== undefined && variables.http_passwd === undefined) {
    variables.http_passwd = variables.http_password
  }
  if (variables['tls.verify'] !== undefined && variables.tls_verify === undefined) {
    variables.tls_verify = variables['tls.verify']
  }
  if (variables.tls_verify !== undefined && variables['tls.verify'] === undefined) {
    variables['tls.verify'] = variables.tls_verify
  }

  if (normalizedType === 'opensearch' && normalizedRuntime === 'fluentd') {
    if (variables.http_user !== undefined && variables.user === undefined) {
      variables.user = variables.http_user
    }
    if (variables.http_passwd !== undefined && variables.password === undefined) {
      variables.password = variables.http_passwd
    }
    if (variables.http_password !== undefined && variables.password === undefined) {
      variables.password = variables.http_password
    }
    if (variables.tls_verify !== undefined && variables.ssl_verify === undefined) {
      variables.ssl_verify = variables.tls_verify
    }
    if (variables.logstash_prefix !== undefined && variables.index_name === undefined) {
      variables.index_name = variables.logstash_prefix
    }
  }

  if (normalizedType === 'http' && normalizedRuntime === 'fluentd' && variables.endpoint === undefined) {
    if (variables.uri && String(variables.uri).startsWith('http')) {
      variables.endpoint = variables.uri
    } else if (variables.host) {
      const portPart = variables.port ? `:${variables.port}` : ''
      const uriPart = variables.uri ? String(variables.uri) : ''
      variables.endpoint = `http://${variables.host}${portPart}${uriPart}`
    }
  }

  return variables
}

export function buildImportedOutputRenderVariables(
  module: any,
  target: any,
  fluentType: string,
): Record<string, any> {
  const targetSettings = parseVariablesMap(target?.settings)
  const instanceVariables = parseVariablesMap(module?.variables)
  return normalizeImportedOutputRenderVariables(target?.target_type, fluentType, {
    ...targetSettings,
    ...instanceVariables,
    output_target_name: target?.name || '',
    output_target_type: target?.target_type || '',
  })
}

export function inferImportedOutputTargetType(module: any): string {
  const plugin = String(module?.output_target_type || module?.detected_plugin || '').trim().toLowerCase()
  if (plugin === 'es' || plugin === 'opensearch' || plugin === 'elasticsearch') return 'opensearch'
  if (plugin === 'loki') return 'loki'
  if (plugin === 'kafka' || plugin === 'rdkafka') return 'kafka'
  if (plugin === 'http') return 'http'
  if (plugin === 's3') return 's3'
  if (plugin === 'stdout') return 'stdout'

  const content = String(module?.content || '').toLowerCase()
  if (content.includes('opensearch') || content.includes('elasticsearch')) return 'opensearch'
  if (content.includes('loki')) return 'loki'
  if (content.includes('kafka')) return 'kafka'
  if (content.includes('stdout')) return 'stdout'
  if (content.includes('s3')) return 's3'
  if (content.includes('http')) return 'http'
  return 'custom'
}

export function buildImportedOutputTargetDraft(module: any): {
  target_type: string
  endpoint: string
  settings: string
} {
  const targetType = inferImportedOutputTargetType(module)
  const rawSettings: Record<string, any> = {
    ...parseVariablesMap(module?.variables),
  }
  if (targetType === 'custom') {
    rawSettings.plugin = module?.detected_plugin || rawSettings.plugin || 'custom_output'
  }
  const settings = pickImportedOutputTargetSettings(rawSettings, targetType)

  let endpoint = ''
  if (rawSettings.uri) {
    endpoint = String(rawSettings.uri)
  } else if (rawSettings.endpoint) {
    endpoint = String(rawSettings.endpoint)
  } else if (rawSettings.host && rawSettings.port) {
    endpoint = `${rawSettings.host}:${rawSettings.port}`
  } else if (rawSettings.host) {
    endpoint = String(rawSettings.host)
  } else if (targetType === 'kafka' && rawSettings.brokers) {
    endpoint = String(rawSettings.brokers)
  } else if (targetType === 's3' && rawSettings.bucket) {
    const path = rawSettings.path ? `/${String(rawSettings.path).replace(/^\/+/, '')}` : ''
    endpoint = `s3://${rawSettings.bucket}${path}`
  } else if (targetType === 'stdout') {
    endpoint = 'stdout'
  }

  return {
    target_type: targetType,
    endpoint,
    settings: JSON.stringify(settings, null, 2),
  }
}

export function importActionBadgeClass(action: string): string {
  return action === 'reuse_existing'
    ? 'bg-success-subtle text-success-emphasis'
    : 'text-bg-light'
}

export function importValidationBadgeClass(verdict: string): string {
  if (verdict === 'equivalent') return 'bg-success-subtle text-success-emphasis'
  if (verdict === 'mostly_equivalent') return 'bg-warning-subtle text-warning-emphasis'
  return 'bg-danger-subtle text-danger-emphasis'
}

export function setImportedModuleAction(module: any, action: string): void {
  if (!module) return
  if (module.module_type === 'output') return
  if (action === 'reuse_existing' && !module.existing_module_id) return
  module.import_action = action
}

// --- t()-dependent ---

export function buildImportedModuleDescription(item: any, t: (k: string) => string): string {
  const details = [t('configs_page.imported_from_existing_config')]
  if (item?.summary) {
    details.push(item.summary)
  }
  return details.join(' · ')
}

export function importActionLabel(action: string, t: (k: string) => string): string {
  if (action === 'reuse_existing') return t('configs_page.import_action_reuse')
  return t('configs_page.import_action_create')
}

export function importValidationLabel(verdict: string, t: (k: string) => string): string {
  if (verdict === 'equivalent') return t('configs_page.import_validation_equivalent')
  if (verdict === 'mostly_equivalent') return t('configs_page.import_validation_mostly_equivalent')
  return t('configs_page.import_validation_needs_review')
}

export function importDestinationMatchLabel(matchType: string, t: (k: string) => string): string {
  if (matchType === 'exact') return t('configs_page.import_destination_match_exact')
  if (matchType === 'created') return t('configs_page.import_destination_match_created')
  return t('configs_page.import_destination_match_type')
}
