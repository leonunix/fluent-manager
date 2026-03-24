import { summarizeOutputTarget } from './output_targets'

function namesFor(modules, type) {
  return Array.from(new Set((modules || [])
    .filter((item) => item.module_type === type)
    .map((item) => item.name)))
}

function summarizeStage(items = []) {
  if (!items.length) return ''
  if (items.length === 1) return items[0]
  if (items.length === 2) return items.join(' + ')
  return `${items[0]} + ${items[1]} + ...`
}

export function buildConfigFlowSummary(modules = [], outputTarget = null) {
  const normalizedTargets = Array.isArray(outputTarget)
    ? outputTarget.filter(Boolean)
    : (outputTarget ? [outputTarget] : [])
  const service = namesFor(modules, 'service')
  const inputs = namesFor(modules, 'input')
  const parsers = namesFor(modules, 'parser')
  const filters = namesFor(modules, 'filter')
  const routes = namesFor(modules, 'route')
  const outputs = namesFor(modules, 'output')
  const processors = [...parsers, ...filters, ...routes]
  const destinations = normalizedTargets.map((target) => summarizeOutputTarget(target))

  const path = []
  const inputLabel = summarizeStage(inputs)
  const processorLabel = summarizeStage(processors)
  const outputLabel = summarizeStage(outputs)
  const destinationNames = normalizedTargets.map((target, index) => target?.name || destinations[index]?.primary).filter(Boolean)
  const destinationLabel = summarizeStage(Array.from(new Set(destinationNames)))
  const destinationChips = []
  for (const [index, target] of normalizedTargets.entries()) {
    const summary = destinations[index]
    if (target?.name && !destinationChips.includes(target.name)) {
      destinationChips.push(target.name)
    }
    for (const chip of summary?.chips || []) {
      if (!destinationChips.includes(chip)) {
        destinationChips.push(chip)
      }
    }
  }

  if (inputLabel) path.push(inputLabel)
  if (processorLabel) path.push(processorLabel)
  if (outputLabel) path.push(outputLabel)
  if (destinationLabel) path.push(destinationLabel)

  return {
    service,
    inputs,
    parsers,
    filters,
    routes,
    outputs,
    processors,
    destinations,
    path,
    inputLabel,
    processorLabel,
    outputLabel,
    destinationLabel,
    destinationChips,
  }
}
