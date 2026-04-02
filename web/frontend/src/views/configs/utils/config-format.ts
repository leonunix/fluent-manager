// Display / badge / label helper functions.
// Pure functions – no reactive state. t-dependent functions accept `t` as last param.

export function runtimeLabel(value: string): string {
  if (value === 'fluentbit') return 'Fluent Bit'
  if (value === 'fluentd') return 'Fluentd'
  if (value === 'shared') return 'Shared'
  return value || '-'
}

export function runtimeBadgeClass(value: string): string {
  if (value === 'fluentbit') return 'bg-info-subtle text-info-emphasis'
  if (value === 'fluentd') return 'bg-warning-subtle text-warning-emphasis'
  return 'bg-secondary-subtle text-secondary-emphasis'
}

export function shortVariables(value: string): string {
  if (!value || value === '{}') return '{}'
  return value.length > 42 ? `${value.slice(0, 42)}...` : value
}

export function parseJSONList(value: string): any[] {
  if (!value) return []
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch {
    return []
  }
}

export function formatJson(value: any): string {
  try {
    return JSON.stringify(value || {}, null, 2)
  } catch {
    return String(value || '{}')
  }
}

export function findingBadgeClass(severity: string): string {
  if (severity === 'error') return 'text-bg-danger'
  if (severity === 'warning') return 'text-bg-warning'
  return 'text-bg-info'
}

export function getProviderErrorMessage(error: any): string {
  return error?.response?.data?.provider_message || ''
}

export function templateAssemblyModules(template: any): any[] {
  return parseJSONList(template?.source_modules)
}

// --- t()-dependent ---

export function templateSourceLabel(sourceType: string, t: (k: string) => string): string {
  return sourceType === 'module_assembly'
    ? t('configs_page.template_source_module_assembly')
    : t('configs_page.template_source_manual')
}

export function wizardGoalLabel(goal: string, t: (k: string) => string): string {
  const labels: Record<string, string> = {
    edge_collection: t('configs_page.goal_edge_collection'),
    central_aggregation: t('configs_page.goal_central_aggregation'),
    custom_pipeline: t('configs_page.goal_custom_pipeline'),
  }
  return labels[goal] || goal
}

export function getErrorMessage(error: any, t: (k: string) => string): string {
  return error?.response?.data?.user_message || error?.response?.data?.error || error?.message || t('common.request_failed')
}
