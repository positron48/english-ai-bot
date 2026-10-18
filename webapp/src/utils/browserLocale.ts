export function getCurrentLocale(): string {
  const stored = localStorage.getItem('locale')
  if (stored && ['ru', 'en', 'es'].includes(stored)) return stored
  const language = navigator.language?.toLowerCase().split('-')[0]
  return language === 'ru' || language === 'es' ? language : 'en'
}
