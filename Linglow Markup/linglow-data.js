// Linglow — App Data
// All static data for the Spanish learning app

const LINGLOW_DATA = {
  user: { name: 'María', streak: 12, words: 128, progress: 62, level: 'A1' },
  dailyRoute: { due: 12, total: 20 },

  districts: [
    { id: 'viajes',     name: 'Distrito de Viajes',      emoji: '✈️',  progress: 3, total: 10, confidence: 68, streak: 12, topicsDone: 46, topicsTotal: 68 },
    { id: 'cafeterias', name: 'Distrito de Cafeterías',   emoji: '☕',  progress: 2, total: 10, confidence: 45, streak: 5,  topicsDone: 20, topicsTotal: 68 },
    { id: 'mercados',   name: 'Distrito de Mercados',     emoji: '🛒',  progress: 1, total: 10, confidence: 25, streak: 3,  topicsDone: 8,  topicsTotal: 68 },
    { id: 'vida',       name: 'Distrito de Vida Diaria',  emoji: '🏠',  progress: 0, total: 10, confidence: 0,  streak: 0,  topicsDone: 0,  topicsTotal: 68 },
    { id: 'parques',    name: 'Distrito de Parques',      emoji: '🌿',  progress: 0, total: 10, confidence: 0,  streak: 0,  topicsDone: 0,  topicsTotal: 68 },
  ],

  buildings: [
    { id: 'repaso',        name: 'Estación de Repaso',     icon: '🏛️', screen: 'grammar' },
    { id: 'grammar',       name: 'Academia de Gramática',  icon: '🎓', screen: 'grammar' },
    { id: 'lectura',       name: 'Avenida de Lectura',     icon: '📚', screen: 'grammar' },
    { id: 'conversation',  name: 'Centro de Conversación', icon: '💬', screen: 'grammar' },
  ],

  districtDetail: {
    // ── Ciudad Luminaria districts ──
    plaza: {
      id: 'plaza', name: 'Plaza Clara',
      desc: 'Центр города. Здесь всё началось — говори, читай, общайся.',
      lumiTip: '¡Ya hablas con confianza aquí! ✨',
      confidence: 82, streak: 12, topicsDone: 54, topicsTotal: 68, level: 'A1',
      practicing: [
        { icon: '📖', title: 'Грамматика', sub: 'ser / estar',     pct: 85 },
        { icon: '🌿', title: 'Слова',      sub: '48 выучено',       pct: 80 },
        { icon: '📚', title: 'Чтение',     sub: 'текст о рынке',    pct: 60 },
        { icon: '💬', title: 'Общение',    sub: 'диалог в кафе',    pct: 55 },
      ],
      tasks: [
        { text: 'Прочитай 2 текста про Plaza Clara', done: 1, total: 2 },
        { text: 'Выучи 10 новых слов из Plaza',       done: 7, total: 10 },
        { text: 'Потренируй диалог в кафе с AI',       done: 0, total: 1 },
      ],
    },
    barrio: {
      id: 'barrio', name: 'Barrio Vivo',
      desc: 'Жилые кварталы — повседневная речь, жизнь города.',
      lumiTip: 'Тут хорошо учить слова для жизни! 🌿',
      confidence: 55, streak: 6, topicsDone: 30, topicsTotal: 68, level: 'A1',
      practicing: [
        { icon: '📖', title: 'Грамматика', sub: 'presente regular',  pct: 65 },
        { icon: '🌿', title: 'Слова',      sub: '32 выучено',         pct: 50 },
        { icon: '📚', title: 'Чтение',     sub: 'письмо соседке',     pct: 30 },
        { icon: '💬', title: 'Общение',    sub: 'знакомство',         pct: 20 },
      ],
      tasks: [
        { text: 'Пройди главу по глаголам -ar',      done: 1, total: 1 },
        { text: 'Повтори 15 слов из Barrio Vivo',     done: 9, total: 15 },
        { text: 'Прочитай текст про жизнь в районе', done: 0, total: 1 },
      ],
    },
    puerta: {
      id: 'puerta', name: 'Puerta de la Chispa',
      desc: 'Ворота в город. Первые шаги, первые слова.',
      lumiTip: 'Продолжи путь отсюда! Ты справишься 🔥',
      confidence: 30, streak: 3, topicsDone: 12, topicsTotal: 68, level: 'A0',
      practicing: [
        { icon: '🌿', title: 'Слова',   sub: '14 выучено',           pct: 35 },
        { icon: '📖', title: 'Азбука', sub: 'алфавит и звуки',       pct: 25 },
        { icon: '💬', title: 'Фразы',  sub: 'приветствия',           pct: 15 },
        { icon: '📚', title: 'Чтение', sub: 'первый текст',          pct: 10 },
      ],
      tasks: [
        { text: 'Выучи базовые приветствия',          done: 2, total: 3 },
        { text: 'Прослушай 5 слов с произношением',   done: 3, total: 5 },
        { text: 'Пройди вводный урок',                done: 0, total: 1 },
      ],
    },
    // ── Legacy fallback ──
    viajes: {
      id: 'viajes',
      name: 'Distrito de Viajes',
      desc: 'Aprende lo esencial para moverte por aeropuertos y medios de transporte.',
      confidence: 68, streak: 12,
      topicsDone: 46, topicsTotal: 68, level: 'A1',
      zones: ['Dirección', 'Estación', 'Billetes', 'Frases útiles', 'Hotel', 'Café de viaje'],
      practicing: [
        { icon: '🔊', title: 'Escuchar',       sub: 'en el aeropuerto', pct: 80 },
        { icon: '💬', title: 'Hablar',         sub: 'en la estación',   pct: 60 },
        { icon: '🎫', title: 'Comprar',        sub: 'billetes',          pct: 40 },
        { icon: '🧳', title: 'El check-in',    sub: 'y el equipaje',     pct: 20 },
      ],
      tasks: [
        { text: 'Completa 3 lecciones sobre los viajes', done: 2, total: 3 },
        { text: 'Practica conversaciones sobre tu próximo viaje', done: 0, total: 1 },
        { text: 'Aprende 10 frases para moverte por la ciudad', done: 7, total: 10 },
      ],
    },
  },

  grammarCategories: [
    { id: 'presente',     name: 'Presente',      icon: '🌱' },
    { id: 'pasado',       name: 'Pasado',        icon: '⌛' },
    { id: 'futuro',       name: 'Futuro',        icon: '🔮' },
    { id: 'pronombres',   name: 'Pronombres',    icon: '👥' },
    { id: 'preposiciones',name: 'Preposiciones', icon: '🪧' },
    { id: 'subjuntivo',   name: 'Subjuntivo',    icon: '⭐' },
  ],

  grammarCategories: [
    { id: 'presente',     name: 'Настоящее время', icon: '🌱' },
    { id: 'pasado',       name: 'Прошедшее',       icon: '⌛' },
    { id: 'articulos',    name: 'Артикли',         icon: '📌' },
    { id: 'pronombres',   name: 'Местоимения',     icon: '👥' },
    { id: 'preposiciones',name: 'Предлоги',        icon: '🪧' },
    { id: 'subjuntivo',   name: 'Subjuntivo',      icon: '⭐' },
  ],

  grammarChapters: {
    presente: [
      { id: 'ar',       num: 1, title: 'Глаголы на -ar',         descRu: 'Как спрягать правильные глаголы на -ar в настоящем времени.', district: 'Plaza Clara', statusRu: 'в процессе', statusColor: 'yellow', pct: 65, icon: 'AR' },
      { id: 'ser-estar',num: 2, title: 'Ser и estar',            descRu: 'Разбираемся в разнице между ser и estar и когда их использовать.', district: 'Plaza Clara', statusRu: 'стабильно',  statusColor: 'green',  pct: 80, icon: '☀️' },
      { id: 'hay-tener',num: 3, title: 'Hay и tener',            descRu: 'Два важных глагола для описания наличия и принадлежности.', district: 'Plaza Clara', statusRu: 'в процессе', statusColor: 'yellow', pct: 30, icon: '🧺' },
      { id: 'preguntas',num: 4, title: 'Вопросительные фразы',  descRu: 'Строим вопросы на испанском правильно и уверенно.', district: 'Plaza Clara', statusRu: 'стабильно',  statusColor: 'green',  pct: 70, icon: '❓' },
    ],
    pasado: [
      { id: 'indef',  num: 1, title: 'Pretérito indefinido', descRu: 'Действия, завершённые в прошлом.', district: 'Barrio Vivo', statusRu: 'не начато', statusColor: 'gray', pct: 0, icon: '📜' },
      { id: 'imperf', num: 2, title: 'Pretérito imperfecto', descRu: 'Описания и привычки в прошлом.', district: 'Barrio Vivo', statusRu: 'не начато', statusColor: 'gray', pct: 0, icon: '🕰️' },
    ],
    articulos:    [],
    pronombres:   [],
    preposiciones:[],
    subjuntivo:   [],
  },

  lesson: {
    id: 'ar', num: 10,
    title: 'El presente: verbos regulares', highlight: '-ar',
    intro: 'En español, los verbos regulares que terminan en -ar siguen un patrón claro en presente. ¡Aprenderlo te ayudará a hablar desde el primer día!',
    rule: 'Para conjugar verbos regulares que terminan en -ar en presente, quita la terminación -ar y añade estas terminaciones:',
    endings: [
      { end: '-o',    pron: 'yo' },
      { end: '-as',   pron: 'tú' },
      { end: '-a',    pron: 'él / ella / usted' },
      { end: '-amos', pron: 'nosotros / as' },
      { end: '-áis',  pron: 'vosotros / as' },
      { end: '-an',   pron: 'ellos / ellas / ustedes' },
    ],
    examples: [
      ['Yo ', 'hablo', ' español todos los días.'],
      ['Tú ', 'hablas', ' muy bien.'],
      ['Ella ', 'habla', ' con su amiga.'],
      ['Nosotros ', 'hablamos', ' en clase.'],
      ['Vosotros ', 'habláis', ' español.'],
      ['Ellos ', 'hablan', ' mucho.'],
    ],
    conjugation: {
      verb: 'hablar',
      rows: [
        ['yo', 'hablo', 'nosotros / nosotras', 'hablamos'],
        ['tú', 'hablas', 'vosotros / vosotras', 'habláis'],
        ['él / ella / usted', 'habla', 'ellos / ellas / ustedes', 'hablan'],
      ],
    },
    tip: 'La mayoría de los verbos que terminan en -ar siguen este patrón. ¡Practica con palabras que uses todos los días!',
  },

  exercises: [
    {
      type: 'choice',
      question: 'el billete',
      context: 'Lo necesitas para viajar en avión o tren.',
      options: ['the map', 'the ticket', 'the passport', 'the suitcase'],
      correct: 1,
      lumiMsg: '¡Tú puedes!\nRecuerda el contexto. 🌿',
      qNum: 5, total: 10,
    },
    {
      type: 'tiles',
      word: 'aeropuerto',
      context: '¿Cómo se escribe…',
      contextSuffix: 'en español?',
      lumiMsg: '¡Tú puedes!',
      qNum: 6, total: 12,
    },
    {
      type: 'write',
      word_en: 'ticket',
      hint: 'se usa para viajar en tren, avión o autobús.',
      correct: 'el billete',
      lumiMsg: '¡Casi!\nEscribe la palabra completa. 🌿',
      qNum: 7, total: 20,
    },
  ],

  progressStats: {
    weekly: [40, 65, 30, 80, 55, 70, 62],
    byDistrict: [
      { name: 'Plaza Clara',         pct: 82 },
      { name: 'Barrio Vivo',         pct: 55 },
      { name: 'Puerta de la Chispa', pct: 30 },
      { name: 'Distrito Alto',       pct: 0  },
      { name: 'Campus de Maestría',  pct: 0  },
    ],
    bySkill: [
      { name: 'Vocabulario', pct: 72 },
      { name: 'Gramática',   pct: 55 },
      { name: 'Lectura',     pct: 48 },
      { name: 'Conversación',pct: 30 },
    ],
  },

  readingTexts: [
    { id: 'dialogo-estacion', title: 'Diálogo en la estación', district: 'plaza', districtName: 'Plaza Clara', districtColor: '#2d6b3a', duration: '5 мин', known: 8, newWords: 3, img: 'uploads/img3.png' },
    { id: 'cafe-plaza',       title: 'Un café en la plaza',   district: 'plaza', districtName: 'Plaza Clara', districtColor: '#2d6b3a', duration: '7 мин', known: 10, newWords: 4, img: 'uploads/img4.png' },
    { id: 'primer-viaje',     title: 'Mi primer viaje',       district: 'puerta', districtName: 'Puerta de la Chispa', districtColor: '#b07830', duration: '6 мин', known: 7, newWords: 5, img: 'uploads/img7.png' },
    { id: 'cartas-sevilla',   title: 'Cartas desde Sevilla',  district: 'travel', districtName: 'Путешествия', districtColor: '#2d5a27', duration: '8 мин', known: 9, newWords: 6, img: 'uploads/img9.png' },
  ],

  readingTextDetail: {
    'dialogo-estacion': {
      title: 'Diálogo en la estación',
      districtName: 'Plaza Clara',
      lumiTip: 'Lee en voz alta para mejorar tu fluidez.',
      img: 'uploads/img3.png',
      lines: [
        { speaker: 'Lucía', es: 'Perdón, ¿este tren va a <b>Distrito Alto</b>?', ru: 'Извините, этот поезд идет в Дистрито Альто?' },
        { speaker: 'Tomás', es: 'Sí, va <b>directo</b>. Sale en cinco minutos.', ru: 'Да, без пересадок. Отправляется через пять минут.' },
        { speaker: 'Lucía', es: '¡Perfecto! ¿Desde qué <b>andén</b> sale?', ru: 'Отлично! С какой платформы отправляется?' },
        { speaker: 'Tomás', es: 'Desde el andén dos, al final del pasillo.', ru: 'Со второй платформы, в конце коридора.' },
        { speaker: 'Lucía', es: 'Muchas gracias.', ru: 'Большое спасибо.' },
        { speaker: 'Tomás', es: 'De nada. Que tengas buen viaje.', ru: 'Не за что. Хорошего путешествия.' },
      ],
      quiz: { q: '¿Desde qué andén sale el tren?', options: ['Desde el andén uno', 'Desde el andén dos', 'Desde el andén tres'], correct: 1 },
    },
  },

  chatScenario: {
    id: 'comprar-billete',
    title: 'Практика ситуации',
    subtitle: 'Comprar un billete',
    district: 'Plaza Clara',
    messages: [
      { role: 'lumi', text: '¡Hola! Soy Lumi, tu compañera de práctica. Hoy vamos a simular que quieres comprar un billete de tren. ¿Listo?' },
      { role: 'user', text: 'Hola, quiero un billete para Sevilla, por favor.' },
      { role: 'lumi', text: 'Perfecto. ¿Para cuándo quieres el billete?' },
      { role: 'user', text: 'Mañana por la mañana.' },
    ],
    correction: {
      original: 'Mañana por la mañana.',
      corrected: 'Mañana por la mañana.',
      note: 'Повторение «mañana» избыточно. Достаточно одного раза для ясности.',
    },
  },
};

window.LINGLOW_DATA = LINGLOW_DATA;
