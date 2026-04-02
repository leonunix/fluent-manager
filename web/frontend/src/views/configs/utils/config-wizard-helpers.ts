// Wizard pure helper functions.
// No reactive state. t-dependent functions accept `t` as last param.

import { uniqueSorted } from './config-text'
import { parseVariablesMap, inferVariableKind, buildWizardVariableDraft, normalizeWizardDraftValues } from './config-variables'

// Module-level sequence counter – shared across all calls in the same app session.
let wizardSequence = 0

export function createWizardInstanceID(prefix: string): string {
  wizardSequence += 1
  return `${prefix}-${wizardSequence}`
}

export function createWizardPipeline(): {
  id: string
  name: string
  input: null
  filters: any[]
  routes: any[]
  outputs: any[]
} {
  return {
    id: createWizardInstanceID('wizard-pipeline'),
    name: '',
    input: null,
    filters: [],
    routes: [],
    outputs: [],
  }
}

export function moduleVariablesForWizard(module: any): Record<string, any> {
  return parseVariablesMap(module?.variables)
}

export function parserNamesProvidedByModule(module: any): string[] {
  const names: string[] = []
  let parserSection = false
  for (const rawLine of String(module?.content || '').split('\n')) {
    const trimmed = rawLine.trim()
    if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith(';')) continue
    if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
      const section = trimmed.slice(1, -1).trim().toUpperCase()
      parserSection = section === 'PARSER' || section === 'MULTILINE_PARSER'
      continue
    }
    if (!parserSection) continue
    const parts = trimmed.split(/\s+/)
    if (parts.length < 2 || parts[0].toLowerCase() !== 'name') continue
    const name = trimmed.slice(parts[0].length).trim()
    if (name && !name.includes('{{')) {
      names.push(name)
    }
  }
  return uniqueSorted(names)
}

export function parserReferencesForInputModule(module: any, draft: Record<string, any> | null = null): string[] {
  const names: string[] = []
  const merged = {
    ...moduleVariablesForWizard(module),
    ...(draft || {}),
  }
  const content = String(module?.content || '')

  // Extract Parser and Parsers directives from the module content.
  for (const rawLine of content.split('\n')) {
    const trimmed = rawLine.trim()
    if (!trimmed || trimmed.startsWith('#') || trimmed.startsWith(';')) continue
    const parts = trimmed.split(/\s+/)
    if (parts.length < 2) continue
    const directive = parts[0].toLowerCase()
    if (directive !== 'parser' && directive !== 'parsers') continue
    const value = trimmed.slice(parts[0].length).trim()
    if (!value || value.includes('{{')) continue
    for (const name of value.split(',').map((s: string) => s.trim()).filter(Boolean)) {
      names.push(name)
    }
  }

  // Also check variable values for parser_name / parser keys.
  for (const [key, value] of Object.entries(merged)) {
    const k = key.toLowerCase()
    if (k !== 'parser_name' && k !== 'parser' && k !== 'parsers') continue
    const raw = String(value || '')
    if (!raw || raw.includes('{{')) continue
    for (const name of raw.split(',').map((s: string) => s.trim()).filter(Boolean)) {
      names.push(name)
    }
  }

  return uniqueSorted(names)
}

export function shouldAutoSyncWizardMatch(currentValue: any, previousTag: string): boolean {
  const current = String(currentValue ?? '').trim()
  return !current || current === '*' || current === '**' || (!!previousTag && current === previousTag)
}

export function buildWizardModuleGroup(
  key: string,
  title: string,
  subtitle: string,
  module: any,
  model: any,
  extraDefaults: Record<string, any> = {},
  section: any = null,
): any | null {
  if (!module || !model) return null
  const defaults = {
    ...moduleVariablesForWizard(module),
    ...(extraDefaults || {}),
  }
  const fields = Object.entries(defaults).map(([fieldKey, value]) => ({
    key: fieldKey,
    kind: inferVariableKind(value),
    description: module.description || '',
  }))
  if (!fields.length) return null
  return {
    key,
    title,
    subtitle,
    sectionKey: section?.key || 'default',
    sectionTitle: section?.title || '',
    sectionKind: section?.kind || 'default',
    sectionRef: section?.ref || '',
    fields,
    model,
  }
}

export function matchingOutputModuleForTarget(
  target: any,
  eligibleModules: any[],
  fluentType: string,
): any | null {
  if (!target) return null
  const matches = eligibleModules.filter((item) =>
    item.module_type === 'output' &&
    item.preset_kind === 'output' &&
    item.preset_key === target.target_type
  )
  return matches.find((item) => item.fluent_type === fluentType) ||
    matches.find((item) => item.fluent_type === 'shared') ||
    null
}

export function buildWizardModuleRef(
  module: any,
  draft: Record<string, any>,
  defaults: Record<string, any> = {},
): any | null {
  if (!module) return null
  const normalized = normalizeWizardDraftValues(draft, {
    ...moduleVariablesForWizard(module),
    ...(defaults || {}),
  })
  if (!Object.keys(normalized).length) {
    return { module_id: module.id }
  }
  return {
    module_id: module.id,
    variables: JSON.stringify(normalized, null, 2),
  }
}

export function ensureWizardModuleDraft(
  _key: string,
  module: any,
  existingDraft: Record<string, any> | null = null,
  extraDefaults: Record<string, any> = {},
): Record<string, string> {
  const defaults = {
    ...moduleVariablesForWizard(module),
    ...(extraDefaults || {}),
  }
  return existingDraft && Object.keys(existingDraft).length
    ? { ...existingDraft }
    : buildWizardVariableDraft(defaults)
}

export function buildOutputTargetModuleRefs(
  targets: any[],
  eligibleModules: any[],
  fluentType: string,
): any[] {
  return targets
    .map((target) => {
      const outputModule = matchingOutputModuleForTarget(target, eligibleModules, fluentType)
      if (!outputModule) return null
      const settings = parseVariablesMap(target.settings)
      const variables = {
        ...settings,
        output_target_name: target.name,
        output_target_type: target.target_type,
      }
      return {
        module_id: outputModule.id,
        variables: JSON.stringify(variables, null, 2),
      }
    })
    .filter(Boolean)
}

// --- t()-dependent ---

export function wizardPipelineDisplayName(
  pipelineOrCard: any,
  index = 0,
  t: (k: string) => string,
): string {
  const pipelineName = String(pipelineOrCard?.name || '').trim()
  return pipelineName || t('configs_page.wizard_pipeline_fallback').replace('{index}', String(index + 1))
}
