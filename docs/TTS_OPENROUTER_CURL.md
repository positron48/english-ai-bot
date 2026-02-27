# Проверка TTS через OpenRouter (curl)

Проверь сам с своим ключом. Подставь `YOUR_OPENROUTER_KEY` из .env (TTS_API_KEY или OPENROUTER_API_KEY).

## Итог по возможностям OpenRouter

- **Отдельного TTS API нет:** у OpenRouter **нет** эндпоинта `/audio/speech` (как у OpenAI). Запрос туда даёт **404**.
- **Озвучка только через chat:** модели с **audio output** (например `openai/gpt-4o-audio-preview`) работают через **`/chat/completions`** с `modalities: ["text", "audio"]` и `stream: true`. В стриме приходят `delta.audio` (base64). Для «озвучить одно слово» нужен жёсткий системный промпт.
- **Модели с audio output (на момент проверки):** `openai/gpt-audio`, `openai/gpt-audio-mini`, `openai/gpt-4o-audio-preview`. Полный список: [openrouter.ai/models](https://openrouter.ai/models) (фильтр output = audio) или `GET https://openrouter.ai/api/v1/models` + jq. Скрипт `./scripts/check_tts_openrouter.sh` выводит список (шаг 3).

---

## 1) OpenRouter — нет endpoint /audio/speech

У OpenRouter **нет** прокси для OpenAI `/audio/speech`. Запрос возвращает **404** (HTML-страница «Not Found»).

```bash
curl -sS -w "\nHTTP %{http_code}\n" -X POST "https://openrouter.ai/api/v1/audio/speech" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_OPENROUTER_KEY" \
  -d '{"model":"tts-1","input":"hello","voice":"alloy","response_format":"mp3"}'
```

Ожидаемо: **404** (или 401 если ключ неверный).

---

## 2) Chat completions с modalities audio

Endpoint есть (401 без ключа, с ключом — ответ от модели). По документации OpenRouter/OpenAI **audio output у gpt-4o-audio не поддерживается** — только audio input (распознавание). Можешь проверить с ключом: если в ответе будет аудио — значит что-то изменилось.

```bash
# Замени YOUR_OPENROUTER_KEY на свой ключ
curl -sS -w "\nHTTP %{http_code}\n" -X POST "https://openrouter.ai/api/v1/chat/completions" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_OPENROUTER_KEY" \
  -d '{
    "model": "openai/gpt-4o-audio-preview",
    "messages": [
      {"role": "system", "content": "You are a pronunciation machine. Say ONLY the exact word the user sent, as audio. One word, no greeting, no pause, no repetition."},
      {"role": "user", "content": "hello"}
    ],
    "modalities": ["text", "audio"],
    "audio": {"voice": "alloy", "format": "pcm16"},
    "stream": true,
    "max_tokens": 150
  }'
```

Если в ответе придут только текст и пустое/нет аудио — через OpenRouter TTS (озвучка одного слова) сейчас недоступен.

---

## 2.1) Одно слово без «минуты тишины»

Чтобы не ждать долгий стрим, можно ограничить ответ и обрезать приём:

1. **`max_tokens`** — задать небольшое значение (например **100–200**). Ответ обрежется по лимиту токенов; для одного слова этого обычно хватает. Подбор: если аудио обрывается — увеличь, если много тишины — уменьши.
2. **Промпт** — ужесточить: *"You are a pronunciation machine. Say ONLY the exact word the user sent, as audio. One word, no greeting, no pause before or after, no repetition."*
3. **Клиент при разборе стрима** — не ждать конец стрима: как только накопили аудио на 1–3 секунды (или первый «кусок» достаточного размера), закрыть чтение и сохранить файл. Так даже при длинном ответе от модели получим только начало (слово) и не будем слушать пустоту.

Пример запроса с `max_tokens`:

```json
{
  "model": "openai/gpt-4o-audio-preview",
  "messages": [...],
  "modalities": ["text", "audio"],
  "audio": {"voice": "alloy", "format": "pcm16"},
  "stream": true,
  "max_tokens": 150
}
```

При реализации в коде: парсить SSE, накапливать `delta.audio` (base64), опционально — прекращать приём после N байт или N секунд (по формату pcm16: 16 kHz mono ≈ 32 000 байт/с).

---

## 3) Список моделей с audio output (через API)

Ключ не обязателен для чтения списка моделей:

```bash
curl -sS "https://openrouter.ai/api/v1/models" | jq -r '
  .data[] | select(.architecture.output_modalities | index("audio")) | .id
'
```

Либо открой в браузере: [openrouter.ai/models](https://openrouter.ai/models) и включи фильтр по output modality «audio».

---

## Итог

- **OpenRouter** не проксирует `/audio/speech`. **TTS только через chat/completions** с моделями, у которых в `output_modalities` есть `audio` (например `openai/gpt-4o-audio-preview`), с `stream: true` и разбором `delta.audio` в стриме.
- Для озвучки одного слова в приложении: **Dictionary API** (бесплатно), **прямой OpenAI** (`api.openai.com` + ключ, `/audio/speech`), либо реализовать в коде путь **chat/completions + stream** с жёстким промптом «только слово» и сбором base64-аудио в файл (формат в ответе — см. документацию модели, возможно pcm16 → конвертация в mp3).

После того как послушаешь/проверишь — напиши, если хочешь вернуть в коде поддержку chat/completions с парсингом стрима для OpenRouter.
