// Search / filter helper functions.
// Pure functions – no reactive state, no i18n.

export function normalizeSearchText(value: any): string {
  return String(value || '').trim().toLowerCase()
}

export function matchesModuleSearch(module: any, keyword: string): boolean {
  const normalizedKeyword = normalizeSearchText(keyword)
  if (!normalizedKeyword) return true

  const haystack = [
    module?.name,
    module?.description,
    module?.module_type,
    module?.fluent_type,
    module?.preset_kind,
    module?.preset_key,
    module?.content,
  ]
    .filter(Boolean)
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalizedKeyword)
}

export function matchesOutputTargetSearch(target: any, keyword: string): boolean {
  const normalizedKeyword = normalizeSearchText(keyword)
  if (!normalizedKeyword) return true

  const haystack = [
    target?.name,
    target?.description,
    target?.target_type,
    target?.endpoint,
    target?.settings,
  ]
    .filter(Boolean)
    .join('\n')
    .toLowerCase()

  return haystack.includes(normalizedKeyword)
}
