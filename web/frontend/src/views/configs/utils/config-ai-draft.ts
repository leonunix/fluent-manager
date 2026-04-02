// AI draft state management utility functions.
// Pure functions – no reactive state, no i18n.

import { contentSimilarityRatio } from './config-text'

export interface AIDraftState {
  active: boolean
  provider: string
  accountName: string
  summary: string
  steps: any[]
  notes: any[]
  reviewItems: any[]
  confirmationItems: Array<{ key: string; label: string; checked: boolean }>
}

export function createAIDraftState(): AIDraftState {
  return {
    active: false,
    provider: '',
    accountName: '',
    summary: '',
    steps: [],
    notes: [],
    reviewItems: [],
    confirmationItems: [],
  }
}

export function resetAIDraftState(state: AIDraftState): void {
  state.active = false
  state.provider = ''
  state.accountName = ''
  state.summary = ''
  state.steps = []
  state.notes = []
  state.reviewItems = []
  state.confirmationItems = []
}

export function buildDraftConfirmationItems(
  labels: string[],
): Array<{ key: string; label: string; checked: boolean }> {
  return labels.map((label, index) => ({
    key: `confirm-${index}`,
    label,
    checked: false,
  }))
}

export function activateAIDraftState(
  state: AIDraftState,
  result: any,
  reviewItems: any[],
  confirmationLabels: string[],
): void {
  state.active = true
  state.provider = result?.provider || ''
  state.accountName = result?.account_name || ''
  state.summary = result?.summary || ''
  state.steps = Array.isArray(result?.assembly_steps) ? result.assembly_steps : []
  state.notes = Array.isArray(result?.notes) ? result.notes : []
  state.reviewItems = reviewItems
  state.confirmationItems = buildDraftConfirmationItems(confirmationLabels)
}

export function areDraftConfirmationsComplete(state: AIDraftState): boolean {
  return !state.active || state.confirmationItems.every((item) => item.checked)
}

// Match each AI-generated module against the existing catalog and assign a default decision.
export function mergeAIModules(generatedModules: any[], existingModules: any[]): any[] {
  return generatedModules.map((mod) => {
    const nameLower = (mod.name || '').toLowerCase()
    const typeLower = (mod.module_type || '').toLowerCase()

    // 1. Exact name + type match.
    const exactMatch = existingModules.find(
      (e) => e.name.toLowerCase() === nameLower && e.module_type.toLowerCase() === typeLower
    )
    if (exactMatch) {
      const ratio = contentSimilarityRatio(mod.content, exactMatch.latest_content || '')
      const decision = ratio >= 0.95 ? 'reuse_existing' : 'update_existing'
      return { ...mod, decision, matchedModule: exactMatch }
    }

    // 2. Content similarity scan across same-type modules (≥70% line overlap).
    const sameType = existingModules.filter((e) => e.module_type.toLowerCase() === typeLower)
    const similarMatch = sameType.find(
      (e) => contentSimilarityRatio(mod.content, e.latest_content || '') >= 0.7
    )
    if (similarMatch) {
      return { ...mod, decision: 'reuse_existing', matchedModule: similarMatch }
    }

    return { ...mod, decision: 'create_new', matchedModule: null }
  })
}
