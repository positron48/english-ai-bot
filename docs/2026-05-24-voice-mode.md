Да, я бы шёл именно в сторону **мультимодального speaking evaluator**, а не “ASR + точное сравнение строки”.

Ключевая идея: **распознавание текста может быть вспомогательным артефактом, но не источником истины**. Пользователь говорит → модель получает аудио + контекст задания + ожидаемый смысл → оценивает, насколько ответ понятен, грамматически нормален, естественен и произнесён достаточно разборчиво.

---

# 1. Архитектура от простого к сложному

## Уровень 1 — асинхронная оценка записи

Это лучший первый шаг.

**UX:**

1. Пользователь видит или слышит задание.
2. Нажимает record.
3. Говорит 3–15 секунд.
4. Получает текстовый фидбек.

**На вход модели:**

* язык: Spanish / English;
* уровень: A0 / A1 / A2 / B1;
* тип задания;
* ожидаемый смысл;
* допустимые варианты ответа;
* аудио пользователя;
* опционально: rough transcript, если есть.

**На выходе:**

```json
{
  "understood_answer": "Quiero un café con leche",
  "meaning_score": 4,
  "grammar_score": 3,
  "pronunciation_score": 3,
  "fluency_score": 3,
  "is_acceptable": true,
  "short_feedback_ru": "Смысл понятен. Лучше добавить артикль: un café.",
  "better_version": "Quiero un café con leche, por favor.",
  "repeat_task": "Повтори: Quiero un café con leche, por favor."
}
```

Это уже можно встроить без real-time диалога.

---

## Уровень 2 — turn-based speaking

Это почти диалог, но технически проще.

**Как работает:**

1. Приложение задаёт вопрос голосом.
2. Пользователь отвечает голосом.
3. Модель оценивает ответ.
4. Модель генерирует следующий вопрос.
5. Приложение озвучивает следующий вопрос через TTS.

То есть это не live voice chat, а **пошаговый диалог**.

Пример:

> App: ¿Qué quieres tomar?
> User: Quiero café.
> Feedback: Хорошо. Более естественно: “Quiero un café, por favor.”
> App: ¿Con leche o solo?

Плюс: дешевле, проще, легче контролировать.
Минус: нет ощущения живого разговора.

---

## Уровень 3 — guided role-play

Это уже полноценные задания-миссии.

Пример:

> Ты в кафе.
> Цель: заказать кофе, спросить цену, поблагодарить.

Модель должна не просто отвечать, а вести пользователя к выполнению целей.

В конце:

```json
{
  "mission_completed": true,
  "goals": [
    {"goal": "ordered_drink", "done": true},
    {"goal": "asked_price", "done": true},
    {"goal": "said_thanks", "done": false}
  ],
  "summary_feedback_ru": "Ты заказал кофе и спросил цену, но не поблагодарил.",
  "useful_phrases": [
    "¿Cuánto cuesta?",
    "Un café, por favor.",
    "Gracias."
  ]
}
```

---

## Уровень 4 — realtime voice dialogue

Это уже продвинутый режим: пользователь говорит, модель перебивает, уточняет, отвечает голосом, держит контекст.

Здесь лучше использовать voice-native модели, а не собирать цепочку:

> ASR → LLM → TTS

Потому что цепочка теряет интонацию, паузы, неуверенность, исправления и даёт больше задержки. OpenAI прямо описывает Realtime API как способ обрабатывать и генерировать аудио напрямую через одну модель/API, чтобы снизить задержку и сохранить речевые нюансы. ([OpenAI][1])

---

# 2. Какие модели смотреть

## OpenAI

### Для MVP: audio/text → text feedback

Подходит схема: **Responses API / Chat API с audio input**. В документации OpenAI вход сообщения может содержать text, image или audio, а audio input передаётся как base64 MP3/WAV. ([OpenAI Developers][2])

Использование:

* оценка произношения “на слух”;
* оценка смысла;
* исправление грамматики;
* рекомендации;
* генерация следующего задания.

Практически я бы начал так:

* дешёвая модель для массовых простых проверок;
* более сильная модель для сложных B1/B2 ответов, свободных пересказов и итогового фидбека;
* отдельная модель/режим для realtime позже.

### Для продвинутого диалога: GPT-Realtime-2

OpenAI в мае 2026 представила GPT-Realtime-2, GPT-Realtime-Translate и GPT-Realtime-Whisper. GPT-Realtime-2 описывается как voice model с GPT-5-class reasoning для live voice interactions; она умеет вести разговор, обрабатывать исправления/перебивания, вызывать tools и держать более длинный контекст. ([OpenAI][3])

Для твоего приложения это кандидат на:

* AI role-play;
* разговорные миссии;
* “экзаменатор”;
* свободный разговор;
* live correction после ответа.

GPT-Realtime-Translate может быть полезен для режима “говори по-русски → услышишь по-испански/английски” или для двуязычных подсказок, а GPT-Realtime-Whisper — если нужен отдельный live transcript. ([OpenAI][3])

---

## Google Gemini

### Для MVP: Gemini audio understanding

Gemini API умеет анализировать аудио и выдавать текстовый ответ: транскрипцию, перевод, ответы по аудио, эмоции и сегменты с timestamps. Но Google отдельно указывает, что generateContent audio не предназначен для realtime transcription; для realtime voice/video надо использовать Live API. ([Google AI for Developers][4])

Хорошо подходит для:

* “оцени мою запись”;
* “что пользователь сказал?”;
* “дай фидбек по ответу”;
* “проверь, понял ли он аудио-вопрос”.

### Для realtime: Gemini Live

У Gemini есть Live-модели для realtime-диалога: Gemini 3.1 Flash Live Preview и Gemini 2.5 Flash Live Preview указаны как модели для low-latency voice-first / bidirectional audio interactions. ([Google AI for Developers][5])

Это альтернатива OpenAI Realtime для advanced режима.

---

## AWS Amazon Nova Sonic

Amazon Nova Sonic — speech-to-speech модель в Bedrock для real-time conversational interactions через bidirectional audio streaming. Документация описывает её как модель, которая обрабатывает и отвечает на речь в реальном времени, без классической цепочки из отдельных STT/LLM/TTS компонентов. ([AWS Documentation][6])

Имеет смысл смотреть, если:

* у тебя уже AWS-инфраструктура;
* нужен enterprise-вариант;
* важны Bedrock, IAM, регионы, корпоративная интеграция.

---

## Claude

Claude хорош как **текстовый reasoning/evaluation слой**, но не как основная audio-модель. Текущая документация Claude говорит, что модели поддерживают text/image input и text output, multilingual и vision; в OpenAI SDK compatibility docs указано, что audio input не поддерживается и будет игнорироваться/stripped. ([Claude API Docs][7])

То есть Claude можно использовать так:

> аудио → отдельная модель получает transcript/notes → Claude оценивает содержание, грамматику, педагогический фидбек.

Но для “модель слушает аудио напрямую” я бы Claude не ставил в первый ряд.

---

## Azure Pronunciation Assessment / SpeechSuper

Это не LLM-диалог, а специализированная оценка произношения. Azure Pronunciation Assessment даёт feedback по accuracy, fluency, completeness и prosody, но Microsoft также указывает, что качество зависит от speech-to-text transcription accuracy и сравнения с reference transcription. ([Microsoft Learn][8])

Я бы использовал это не вместо LLM, а как **дополнительный слой** для scripted pronunciation drills:

> пользователь должен сказать конкретную фразу → Azure/SpeechSuper даёт phoneme/fluency metrics → LLM превращает это в человеческий фидбек.

---

# 3. Что я бы выбрал по этапам

## Этап 1 — Speaking Evaluator

**Цель:** быстро добавить ценность без realtime.

Режимы:

1. “Скажи фразу”
2. “Ответь на вопрос”
3. “Опиши картинку/ситуацию”
4. “Перескажи текст”
5. “Повтори улучшенный вариант”

**Технически:**

```text
Frontend:
record audio → upload

Backend:
task_context + audio → multimodal model → structured JSON feedback

App:
показывает фидбек + better version + кнопку "повторить"
```

**Главное:** не ставить пользователю “ты сказал не то”, если аудио неоднозначное. Лучше:

> Я понял это как: “Quiero café.”
> Смысл правильный. Лучше: “Quiero un café, por favor.”

---

## Этап 2 — Controlled speaking cards

Это задания с ограниченной областью ответа.

### Пример 1 — Translate and say

```text
Скажи по-испански:
“Я хочу кофе.”
```

Ожидаемые варианты:

```text
Quiero café.
Quiero un café.
Me gustaría un café.
```

Фидбек:

```text
Смысл понятен. Лучше: “Quiero un café.”
В испанском здесь естественно добавить артикль “un”.
```

---

### Пример 2 — Answer the question

```text
Listen:
¿Dónde vives?

Answer in Spanish.
```

Пользователь:

```text
Vivo en Berlín.
```

Фидбек:

```text
Отлично. Ответ короткий и правильный.
Можно добавить: “Vivo en Berlín, Alemania.”
```

---

### Пример 3 — Personal response

```text
Answer in English:
What do you usually do in the evening?
```

Пользователь:

```text
I usually read or watch videos.
```

Фидбек:

```text
Good answer. Natural and clear.
More advanced version: “I usually read a book or watch videos before going to bed.”
```

---

## Этап 3 — Listen → speak

Тут пользователь не видит текст вопроса, только слышит.

**Задание:**

> Пользователь слышит:
> “What time do you usually wake up?”

Потом отвечает голосом.

Это лучше тренирует реальную коммуникацию, потому что пользователь должен сначала понять вопрос на слух.

**Оценивать надо два навыка:**

1. понял ли вопрос;
2. смог ли ответить.

Пример фидбека:

```text
Ты понял вопрос правильно и ответил по смыслу.
Грамматически лучше: “I usually wake up at seven.”
Не забудь “at” перед временем.
```

---

## Этап 4 — Repair loop

Очень сильная механика.

**Цикл:**

1. Пользователь отвечает.
2. Модель исправляет.
3. Пользователь повторяет исправленную версию.
4. Модель проверяет повтор.

Пример:

Пользователь:

```text
I want go home.
```

Фидбек:

```text
Смысл понятен, но нужно “to”:
“I want to go home.”
Повтори эту фразу.
```

Пользователь повторяет.

Фидбек:

```text
Теперь правильно. Попробуй ещё раз чуть быстрее.
```

Это полезнее, чем просто показать ошибку.

---

## Этап 5 — Диалоги по ролям, но пошаговые

Начать лучше не с open-ended AI-чата, а с управляемых сценариев.

### Сценарий: Café, A1

```text
Goal:
Order a drink and ask the price.

Assistant role:
Waiter.

User role:
Customer.

Required user intents:
- greet
- order coffee
- ask price
- say thanks
```

Диалог:

```text
App: Hola. ¿Qué quieres tomar?
User: Quiero un café.
App: Claro. ¿Algo más?
User: ¿Cuánto cuesta?
App: Dos euros.
User: Gracias.
```

Итог:

```text
Mission completed: 4/4
Good phrase: “¿Cuánto cuesta?”
Improve: “Quiero un café, por favor” sounds more polite.
```

---

# 4. Примеры режимов для приложения

## Режим 1. “Скажи”

Для A0/A1.

```text
Скажи по-английски:
“Мне нужен билет.”
```

Ожидаем:

```text
I need a ticket.
```

Фидбек:

```text
Смысл понятен. Не забудь артикль: “a ticket”.
```

---

## Режим 2. “Ответь”

```text
Question:
Do you like coffee?

Answer with a full sentence.
```

Хороший ответ:

```text
Yes, I like coffee.
No, I don't like coffee.
```

---

## Режим 3. “Выбери и скажи”

Для совсем слабых пользователей.

```text
You are in a café. Choose what you want and say it:

☕ coffee
🍵 tea
💧 water
```

Пользователь говорит:

```text
Quiero té.
```

---

## Режим 4. “Скажи без подсказки”

Сначала пользователь видит пример, потом пример исчезает.

```text
Round 1:
Quiero un café, por favor.

Round 2:
Скажи это сам без текста.
```

---

## Режим 5. “Перескажи”

После текста или аудио.

```text
Listen to the short story.
Now say what happened in 1–2 sentences.
```

Фидбек:

```text
Ты правильно передал общий смысл.
Лучше использовать past tense:
“He went to the shop” вместо “He go to shop”.
```

---

## Режим 6. “Маленькое интервью”

```text
Topic: work

Questions:
1. What do you do?
2. Do you like your job?
3. What do you usually do at work?
```

После 3 ответов:

```text
Summary:
Ты можешь говорить о работе простыми фразами.
Главная ошибка: пропуск do/does в вопросах и отрицаниях.
Новая фраза: “I work as a developer.”
```

---

## Режим 7. “Миссия”

```text
Ты в отеле.
Тебе нужно:
1. сказать, что у тебя бронь;
2. назвать имя;
3. спросить про Wi-Fi;
4. спросить время завтрака.
```

Это уже ближе к реальному speaking.

---

## Режим 8. “Экзаменатор”

Для B1+.

```text
You will speak for 60 seconds.
Topic: Describe a city you like.
```

Оценка:

```text
Content: good
Grammar: okay
Fluency: needs work
Vocabulary: good

Main correction:
“I have been in Valencia” → “I have been to Valencia.”
```

---

## Режим 9. “Разговор с ограничением”

Хороший формат для продвинутых.

```text
Speak only in Spanish.
The assistant may help, but only after you try.
If you switch to Russian, the assistant gives a hint in Spanish first.
```

---

# 5. Как не превратить это в борьбу с распознаванием

Главные правила:

## 1. Оценивать смысл, а не точное совпадение

Плохо:

```text
Expected: I need a ticket.
Heard: I need ticket.
Result: wrong.
```

Хорошо:

```text
Смысл понятен. Нужен артикль:
“I need a ticket.”
```

---

## 2. Всегда показывать “что модель поняла”

```text
Я понял твой ответ так:
“Quiero café.”

Если это не то, что ты сказал, попробуй ещё раз медленнее.
```

Это снимает фрустрацию.

---

## 3. Ввести статус “audio unclear”

Не каждая ошибка — ошибка ученика.

```json
{
  "audio_quality": "unclear",
  "is_acceptable": null,
  "feedback": "Запись не очень разборчивая. Попробуй сказать ближе к микрофону."
}
```

---

## 4. Не давать псевдоточность

Не надо:

```text
Pronunciation: 73%
```

Лучше:

```text
Произношение в целом понятно.
Сложное место: слово “three”.
Попробуй не произносить его как “tree”.
```

---

## 5. Разделять типы ошибок

```json
{
  "meaning": "понятно",
  "grammar": "есть ошибка",
  "pronunciation": "понятно",
  "fluency": "медленно, но нормально"
}
```

Так пользователь понимает, что именно улучшать.

---

# 6. Рекомендуемый порядок внедрения

## Версия 1

**Speaking cards с мультимодальной оценкой.**

Минимальный набор:

* запись аудио;
* отправка audio + task context в модель;
* JSON-фидбек;
* better version;
* кнопка “повторить”.

Это уже даст ощущение настоящей speaking-практики.

---

## Версия 2

**Listen → answer.**

Пользователь слышит вопрос, не видит текст, отвечает голосом.

Это добавит связку:

> listening comprehension → speaking

---

## Версия 3

**Repair loop.**

Пользователь не просто получает исправление, а повторяет правильный вариант.

Это сильно повышает обучающую ценность.

---

## Версия 4

**Пошаговые role-play миссии.**

Без realtime, но с состоянием диалога:

```text
assistant asks → user answers → evaluator checks → next step
```

---

## Версия 5

**Realtime AI role-play.**

Тут уже подключать GPT-Realtime-2 / Gemini Live / Nova Sonic.

Сценарии:

* café;
* hotel;
* airport;
* doctor;
* work meeting;
* apartment viewing;
* job interview;
* small talk;
* immigration office;
* phone call.

---

# 7. Моя практическая рекомендация

Я бы сделал так:

**MVP:**
OpenAI или Gemini audio input → text feedback. Без realtime. Только speaking cards и short answers.

**Следующий шаг:**
turn-based role-play. Модель оценивает каждый ответ и генерирует следующую реплику.

**Потом:**
native realtime voice model для живых диалогов.

**Не начинал бы сразу с realtime**, потому что сложность резко возрастает: latency, interruptions, state management, safety, стоимость, запись с микрофона, шумы, reconnect, mobile edge cases.

Самый сильный первый режим для твоего приложения:

```text
Слушай вопрос → ответь голосом → получи короткий фидбек → повтори улучшенный вариант.
```

Это закрывает speaking, listening, grammar и pronunciation без превращения в “угадай, что распознал ASR”.

[1]: https://openai.com/index/introducing-gpt-realtime/ "Introducing gpt-realtime and Realtime API updates for production voice agents | OpenAI"
[2]: https://developers.openai.com/api/reference/resources/responses/subresources/input_tokens/methods/count/ "Get input token counts | OpenAI API Reference"
[3]: https://openai.com/index/advancing-voice-intelligence-with-new-models-in-the-api/ "Advancing voice intelligence with new models in the API | OpenAI"
[4]: https://ai.google.dev/gemini-api/docs/audio "Gemini generateContent API  |  Google AI for Developers"
[5]: https://ai.google.dev/gemini-api/docs/models "Models  |  Gemini API  |  Google AI for Developers"
[6]: https://docs.aws.amazon.com/nova/latest/userguide/speech.html?utm_source=chatgpt.com "Using the Amazon Nova Sonic Speech-to-Speech model"
[7]: https://docs.anthropic.com/en/docs/about-claude/models "Models overview - Claude API Docs"
[8]: https://learn.microsoft.com/en-us/azure/ai-services/speech-service/how-to-pronunciation-assessment?utm_source=chatgpt.com "Use pronunciation assessment - Foundry Tools"
