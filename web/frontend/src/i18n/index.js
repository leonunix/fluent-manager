import { ref, computed } from 'vue'
import zh from './zh'
import en from './en'
import ja from './ja'

const messages = { zh, en, ja }
const locale = ref(localStorage.getItem('fm_locale') || 'zh')

export function useI18n() {
  function t(key) {
    const keys = key.split('.')
    let val = messages[locale.value]
    for (const k of keys) {
      val = val?.[k]
    }
    if (val !== undefined) return val
    // Fallback to zh
    val = messages.zh
    for (const k of keys) {
      val = val?.[k]
    }
    return val !== undefined ? val : key
  }

  function setLocale(l) {
    locale.value = l
    localStorage.setItem('fm_locale', l)
  }

  return {
    t,
    locale: computed(() => locale.value),
    setLocale,
    availableLocales: ['zh', 'en', 'ja'],
    localeNames: { zh: '中文', en: 'English', ja: '日本語' },
  }
}
