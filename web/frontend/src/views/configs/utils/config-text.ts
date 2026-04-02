// General string / text utility functions.
// Pure functions – no reactive state, no i18n.

export const WIZARD_STAGE_PAGE_SIZE = 6

export function normalizeName(value: any): string {
  return String(value || '').trim().toLowerCase()
}

export function generateUniqueDraftName(
  baseName: string,
  existingNames: string[],
  fallbackPrefix = 'ai-draft',
): string {
  const normalizedExisting = new Set(existingNames.map((item) => normalizeName(item)).filter(Boolean))
  const seed = String(baseName || '').trim() || fallbackPrefix
  if (!normalizedExisting.has(normalizeName(seed))) {
    return seed
  }

  let index = 2
  let candidate = `${seed}-${index}`
  while (normalizedExisting.has(normalizeName(candidate))) {
    index += 1
    candidate = `${seed}-${index}`
  }
  return candidate
}

export function countNonEmptyLines(content: string): number {
  return String(content || '')
    .split('\n')
    .map((line) => line.trim())
    .filter(Boolean).length
}

export function extractTemplatePlaceholders(content: string): string[] {
  const matches = String(content || '').match(/{{\s*\.[^}]+}}/g)
  return matches || []
}

export function uniqueSorted(values: any[]): any[] {
  return Array.from(new Set(values.filter(Boolean))).sort()
}

export function diffKeys(
  current: string[],
  previous: string[],
): { added: string[]; removed: string[] } {
  const currentSet = new Set(current)
  const previousSet = new Set(previous)
  return {
    added: current.filter((item) => !previousSet.has(item)),
    removed: previous.filter((item) => !currentSet.has(item)),
  }
}

export function sortObjectKeys(value: any): any {
  if (Array.isArray(value)) {
    return value.map((item) => sortObjectKeys(item))
  }
  if (!value || typeof value !== 'object') {
    return value
  }
  return Object.keys(value)
    .sort()
    .reduce((acc: Record<string, any>, key) => {
      acc[key] = sortObjectKeys(value[key])
      return acc
    }, {})
}

export function stringifySortedObject(value: any): string {
  return JSON.stringify(sortObjectKeys(value || {}))
}

// Normalise a single line/segment for comparison: collapse inline whitespace, lowercase.
export function normalizeLine(s: string): string {
  return String(s || '').replace(/[ \t]+/g, ' ').trim().toLowerCase()
}

// Split raw content into comparable segments (newline or semicolon delimited), normalising each.
export function splitContentLines(raw: string): string[] {
  return String(raw || '')
    .split(/[;\n]/)
    .map(normalizeLine)
    .filter(Boolean)
}

// Ratio of shared segments between two raw content strings.
export function contentSimilarityRatio(rawA: string, rawB: string): number {
  const aLines = splitContentLines(rawA)
  const bLines = splitContentLines(rawB)
  if (!aLines.length || !bLines.length) return 0
  const bSet = new Set(bLines)
  const matching = aLines.filter((l) => bSet.has(l)).length
  return matching / Math.max(aLines.length, bLines.length)
}

export function paginateItems(
  items: any[],
  page: number,
  pageSize = WIZARD_STAGE_PAGE_SIZE,
): { items: any[]; totalPages: number; currentPage: number; total: number } {
  const totalPages = Math.max(1, Math.ceil(items.length / pageSize))
  const currentPage = Math.min(Math.max(Number(page) || 1, 1), totalPages)
  const start = (currentPage - 1) * pageSize
  return {
    items: items.slice(start, start + pageSize),
    totalPages,
    currentPage,
    total: items.length,
  }
}

// --- t()-dependent ---

export function summarizeChangedKeys(
  added: string[],
  removed: string[],
  t: (k: string) => string,
): string {
  const parts: string[] = []
  if (added.length) {
    parts.push(t('configs_page.ai_draft_diff_added').replace('{items}', added.join(', ')))
  }
  if (removed.length) {
    parts.push(t('configs_page.ai_draft_diff_removed').replace('{items}', removed.join(', ')))
  }
  return parts.join('；')
}
