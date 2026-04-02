// Variable parsing and manipulation functions.
// Pure functions – no reactive state, no i18n.

export function parseVariablesMap(value: any): Record<string, any> {
  if (!value) return {}
  try {
    const parsed = JSON.parse(value)
    return parsed && typeof parsed === 'object' && !Array.isArray(parsed) ? parsed : {}
  } catch {
    return {}
  }
}

// Throws a plain Error if the JSON is not an object – callers handle user-facing message.
export function parseVariablesMapStrict(value: any): Record<string, any> {
  const raw = String(value || '').trim()
  if (!raw) return {}
  const parsed = JSON.parse(raw)
  if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) {
    throw new Error('variable_json_object_required')
  }
  return parsed
}

export function inferVariableKind(value: any): string {
  if (typeof value === 'boolean') return 'boolean'
  if (value && typeof value === 'object') return 'json'
  return 'text'
}

export function stringifyVariableValue(value: any): string {
  if (typeof value === 'boolean') return value ? 'true' : 'false'
  if (value && typeof value === 'object') return JSON.stringify(value, null, 2)
  if (value === undefined || value === null) return ''
  return String(value)
}

export function buildWizardVariableDraft(defaults: Record<string, any>): Record<string, string> {
  const draft: Record<string, string> = {}
  for (const [key, value] of Object.entries(defaults || {})) {
    draft[key] = stringifyVariableValue(value)
  }
  return draft
}

export function normalizeWizardVariableValue(value: any, kind: string): any {
  if (kind === 'boolean') {
    return value === true || value === 'true'
  }
  if (kind === 'json') {
    if (!String(value || '').trim()) return {}
    try {
      return JSON.parse(value)
    } catch {
      return value
    }
  }
  return value
}

export function normalizeWizardDraftValues(
  draft: Record<string, any>,
  defaults: Record<string, any>,
): Record<string, any> {
  const normalized: Record<string, any> = {}
  const mergedDefaults = { ...(defaults || {}) }
  const keys = new Set([...Object.keys(mergedDefaults), ...Object.keys(draft || {})])
  for (const key of keys) {
    normalized[key] = normalizeWizardVariableValue(draft?.[key], inferVariableKind(mergedDefaults[key]))
  }
  return normalized
}

export function inferModuleVariableRowType(value: any): string {
  if (typeof value === 'boolean') return 'boolean'
  if (typeof value === 'number') return 'number'
  if (value && typeof value === 'object') return 'json'
  return 'string'
}

export function parseModuleVariableRowValue(row: { type: string; value: any }): any {
  if (row.type === 'boolean') {
    return row.value === true || row.value === 'true'
  }
  if (row.type === 'number') {
    const trimmed = String(row.value ?? '').trim()
    if (trimmed === '') return 0
    const parsed = Number(trimmed)
    if (Number.isNaN(parsed)) {
      throw new Error('invalid number')
    }
    return parsed
  }
  if (row.type === 'json') {
    const trimmed = String(row.value ?? '').trim()
    if (!trimmed) return {}
    return JSON.parse(trimmed)
  }
  return String(row.value ?? '')
}

export function buildModuleVariableRows(
  raw: any,
): Array<{ key: string; type: string; value: any }> {
  const parsed = parseVariablesMapStrict(raw)
  const entries = Object.entries(parsed)
  if (!entries.length) {
    return [{ key: '', type: 'string', value: '' }]
  }
  return entries.map(([key, value]) => ({
    key,
    type: inferModuleVariableRowType(value),
    value: stringifyVariableValue(value),
  }))
}
