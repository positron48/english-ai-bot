import ru from '../../locales/ru.json'
import en from '../../locales/en.json'
import es from '../../locales/es.json'
import { getCurrentLocale } from '../../utils/browserLocale'
const messages = { ru, en, es }
export function journeyMessage(key: keyof typeof ru.dailyJourney.messages): string {
  const locale = getCurrentLocale() as keyof typeof messages
  return (messages[locale] || messages.en).dailyJourney.messages[key]
}
