# Linglow: идеи NPC и цепочек квестов для conversation mode

Документ фиксирует контент-план для будущих миграций `conversation_scenarios` и `conversation_tasks`.
Цель: заранее набросать NPC, последовательные истории и цепочки квестов отдельно для курсов `en_ru` и `es_ru`, чтобы позже можно было выбрать конкретного персонажа и развернуть его в SQL seed/import.

## Статус реализации

- Сделано: в списке NPC показывается специализация рядом с именем на языке курса (`Mara, barista` для `en_ru`, `Mara, barista de cafetería` для `es_ru`).
- Сделано: продолжение цепочки Mara добавлено в миграции `internal/database/migrations/000049_extend_mara_cafe_chain.sql`.
- Сделано: продолжения цепочек Sam и Officer Park добавлены в миграции `internal/database/migrations/000050_extend_sam_park_chains.sql`.
- Не начато: новые NPC для English course и Spanish course.

## Текущая модель conversation mode

NPC-квесты хранятся в таблицах `conversation_scenarios` и `conversation_tasks`.
Каждый сценарий получает backing `learning_items` с `item_type='speaking_task'`, `source_kind='conversation_scenario'`, `source_id=<scenario.code>`.

Цепочки строятся через:

- `npc_code` - общий код NPC для группировки сценариев;
- `prerequisite_code` - `code` предыдущего сценария, который должен быть завершён;
- `sort_order` - порядок внутри цепочки;
- `cefr_level` - уровень сценария, совпадающий с `districts.level_code`;
- `course_id` - отдельная копия сценария для `en_ru` и `es_ru`.

Текущие засеянные и запланированные цепочки:

- A0, Mara, `mara_barista`: `cafe_order_coffee` -> `cafe_order_pastry` -> `cafe_choose_seat` -> `cafe_wrong_order` -> `cafe_favorite_drink` -> `cafe_busy_morning` -> `cafe_mara_dream` -> `cafe_free_chat`;
- A1, Sam, `sam_shop`: `shop_buy_water` -> `shop_ask_directions` -> `shop_return_item` -> `shop_find_gift` -> `shop_compare_items` -> `shop_lost_receipt` -> `shop_mystery_object` -> `shop_note_in_bottle` -> `shop_secret_shelf`;
- A2, Officer Park, `park_police`: `police_report_lost` -> `police_describe_person` -> `police_ask_next_steps` -> `police_missing_pet` -> `police_timeline` -> `police_witness_statement` -> `police_noise_complaint` -> `police_connect_clues` -> `police_case_closed`.

Практические правила для будущего контента:

- пользовательские заголовки и названия задач пишутся на русском;
- `npc_persona`, `scene_setup`, `completion_criteria` пишутся как model-facing инструкции на английском;
- `completion_criteria` должны быть семантическими и языконейтральными, без конкретных английских или испанских фраз;
- NPC говорит на target language курса: English для `en_ru`, Spanish для `es_ru`;
- A0-A2 требуют очень коротких и простых реплик, B1-B2 допускают объяснения, аргументацию и пересказ, C1 - нюансы, абстрактные темы и моральные дилеммы.

## Общие продолжения текущих NPC

Эти цепочки можно добавить в оба курса, потому что персонажи уже знакомы пользователю.

### Mara, бариста, A0

Статус: реализовано в `internal/database/migrations/000049_extend_mara_cafe_chain.sql`.

История: Мара постепенно узнаёт ученика как постоянного гостя, просит помочь с простыми вещами в кафе, а потом делится мечтой открыть маленькое вечернее кафе с историями.

Квесты:

1. Сделано: `cafe_choose_seat` - попросить место у окна или столик.
2. Сделано: `cafe_wrong_order` - вежливо сказать, что заказ не тот.
3. Сделано: `cafe_favorite_drink` - рассказать, какой напиток нравится.
4. Сделано: `cafe_busy_morning` - помочь Маре в загруженное утро.
5. Сделано: `cafe_mara_dream` - спросить о мечте Мары и рассказать о своей.
6. Сделано: `cafe_free_chat` - существующий free chat передвинут в конец цепочки.

### Sam, продавец, A1

Статус: реализовано в `internal/database/migrations/000050_extend_sam_park_chains.sql`.

История: Сэм кажется обычным продавцом, но собирает странные вещи, которые люди забывают в магазине. Из этих вещей складывается маленькая городская тайна.

Квесты:

1. Сделано: `shop_find_gift` - выбрать простой подарок.
2. Сделано: `shop_compare_items` - сравнить два товара.
3. Сделано: `shop_lost_receipt` - объяснить, что чек потерян.
4. Сделано: `shop_mystery_object` - описать странный предмет из коробки забытых вещей.
5. Сделано: `shop_note_in_bottle` - прочитать или обсудить записку, найденную в бутылке.
6. Сделано: `shop_secret_shelf` - free chat: Сэм показывает "секретную полку" города.

### Officer Park, полицейский, A2

Статус: реализовано в `internal/database/migrations/000050_extend_sam_park_chains.sql`.

История: Парк ведёт серию небольших, нестрашных городских дел: потерянные вещи, странные свидетели, ночные звуки. Игрок становится добровольным помощником.

Квесты:

1. Сделано: `police_missing_pet` - сообщить о пропавшем питомце.
2. Сделано: `police_timeline` - рассказать, что произошло сначала и потом.
3. Сделано: `police_witness_statement` - дать простое свидетельское описание.
4. Сделано: `police_noise_complaint` - описать ночной шум.
5. Сделано: `police_connect_clues` - сопоставить две подсказки.
6. Сделано: `police_case_closed` - спросить, чем всё закончилось.

## English course: новые NPC

### Dr. June Vale, психолог, A2-B1

История: спокойный психолог в городском центре. Сначала учит говорить о себе простыми словами, потом помогает герою разобраться с привычками, страхами и целями.

Квесты:

1. A2 `therapy_first_visit` - назвать имя, настроение, причину визита.
2. A2 `therapy_daily_routine` - описать обычный день.
3. A2 `therapy_small_worry` - рассказать о небольшой тревоге.
4. B1 `therapy_old_habit` - объяснить привычку и почему её трудно изменить.
5. B1 `therapy_set_goal` - сформулировать цель на неделю.
6. B1 `therapy_letter_to_self` - обсудить письмо себе из будущего.

### Professor Rowan Pike, университетский преподаватель, B1-B2

История: профессор ведёт курс "городские легенды и язык". Он просит студента помогать собирать истории жителей, постепенно открывая, что некоторые легенды похожи на правду.

Квесты:

1. B1 `university_join_class` - представиться и объяснить интерес к курсу.
2. B1 `university_ask_assignment` - уточнить задание.
3. B1 `university_present_legend` - кратко пересказать городскую легенду.
4. B2 `university_debate_sources` - обсудить, можно ли доверять источникам.
5. B2 `university_interview_plan` - составить план интервью.
6. B2 `university_final_theory` - защитить свою версию тайны.

### Nora Finch, аптекарь, A1-A2

История: аптекарь знает всех жителей района и часто слышит странные симптомы, жалобы и слухи. Квесты тренируют здоровье, просьбы, инструкции.

Квесты:

1. A1 `pharmacy_buy_medicine` - попросить простое лекарство.
2. A1 `pharmacy_ask_dosage` - спросить, как принимать.
3. A2 `pharmacy_describe_symptoms` - описать симптомы.
4. A2 `pharmacy_allergy_warning` - сказать об аллергии или ограничении.
5. A2 `pharmacy_help_neighbor` - купить лекарство для другого человека.
6. A2 `pharmacy_strange_prescription` - обсудить странный рецепт без мистики и опасных советов.

### Dr. Elias Stone, врач, A2-B1

История: врач в маленькой клинике. Сначала бытовая медицина, потом пациент помогает ему понять цепочку необычных, но безопасных городских случаев.

Квесты:

1. A2 `clinic_make_appointment` - записаться на приём.
2. A2 `clinic_describe_pain` - описать боль или самочувствие.
3. A2 `clinic_answer_questions` - ответить на вопросы о сне, еде, температуре.
4. B1 `clinic_explain_history` - рассказать, когда всё началось.
5. B1 `clinic_follow_advice` - обсудить рекомендации врача.
6. B1 `clinic_city_pattern` - заметить связь между жалобами жителей.

### Mr. Bell, городской сумасшедший у часов, A1-B1

История: странный, но добрый человек у старых часов. Говорит загадками, но постепенно становится понятно, что он просто очень одинок и хранит память о прошлом города.

Квесты:

1. A1 `clockman_first_warning` - понять простое предупреждение.
2. A1 `clockman_ask_time` - спросить время и дорогу.
3. A2 `clockman_lost_day` - выслушать рассказ о "потерянном дне".
4. A2 `clockman_find_photo` - описать старую фотографию.
5. B1 `clockman_memory_story` - обсудить воспоминание и чувства.
6. B1 `clockman_last_chime` - помочь ему решить, что делать дальше.

### Iris Lane, прохожая-детектив, B1

Рефрен: нуарная сыщица в плаще, без прямых отсылок к конкретным героям.

История: она постоянно "случайно" встречается на улицах и просит помочь с мелкими наблюдениями.

Квесты:

1. `detective_street_question` - ответить, кого или что видел на улице.
2. `detective_describe_place` - описать место.
3. `detective_follow_clue` - объяснить маршрут по подсказкам.
4. `detective_compare_versions` - сравнить две версии событий.
5. `detective_hidden_motive` - предположить мотив.
6. `detective_rainy_confession` - обсудить финальный поворот.

### Captain Grey, водитель ночного автобуса, A2-B1

История: водитель знает маршруты и людей. Ночной автобус становится местом маленьких историй пассажиров.

Квесты:

1. A2 `bus_buy_ticket` - купить билет и назвать остановку.
2. A2 `bus_missed_stop` - объяснить, что проехал остановку.
3. A2 `bus_ask_route` - спросить, как пересесть.
4. B1 `bus_talk_passenger` - пересказать историю пассажира.
5. B1 `bus_last_route` - обсудить последний рейс и странного пассажира.
6. B1 `bus_city_map_secret` - понять, почему маршрут изменился.

### Lila Moon, уличная музыкантка, A1-B1

История: играет на площади, собирает песни прохожих и ищет мелодию, которую напевала её бабушка.

Квесты:

1. A1 `musician_give_coin` - похвалить музыку и дать монету.
2. A1 `musician_request_song` - попросить песню.
3. A2 `musician_describe_mood` - описать настроение песни.
4. A2 `musician_lost_melody` - помочь вспомнить часть мелодии словами.
5. B1 `musician_family_story` - обсудить семейную историю.
6. B1 `musician_square_concert` - помочь подготовить маленький концерт.

### Ada Quill, библиотекарь, B1-C1

История: библиотекарь охраняет "тихую полку", где книги появляются сами. Хорошо для пересказа, гипотез, абстрактной речи.

Квесты:

1. B1 `library_get_card` - оформить читательский билет.
2. B1 `library_find_book` - описать нужную книгу.
3. B1 `library_summarize_chapter` - пересказать главу.
4. B2 `library_argue_interpretation` - обсудить смысл текста.
5. B2 `library_unreliable_narrator` - понять ненадёжного рассказчика.
6. C1 `library_book_that_reads_you` - философски обсудить книгу, которая "выбирает" читателя.

### Victor North, бывший герой без имени, B2-C1

Рефрен: уставший супергерой или защитник города, но без костюмов, эмблем и имён.

История: он больше не спасает мир, а учится жить обычной жизнью. Игрок помогает ему говорить честно.

Квесты:

1. B2 `retired_hero_small_errand` - помочь с обычной бытовой задачей.
2. B2 `retired_hero_public_attention` - обсудить нежелательное внимание.
3. B2 `retired_hero_moral_choice` - выбрать между двумя плохими решениями.
4. C1 `retired_hero_cost_of_fame` - обсудить цену славы.
5. C1 `retired_hero_apology` - подготовить сложное извинение.
6. C1 `retired_hero_new_identity` - решить, кем он хочет быть теперь.

## Spanish course: новые NPC

### Inés Robles, психолог, A2-B1

История: тёплая психологиня из районного центра. Больше испанского колорита: семья, привычки, личные границы, планы.

Квесты:

1. A2 `therapy_ines_first_session` - рассказать, как себя чувствуешь.
2. A2 `therapy_ines_family_balance` - просто описать семью или близких.
3. A2 `therapy_ines_bad_day` - рассказать о плохом дне.
4. B1 `therapy_ines_say_no` - потренироваться вежливо отказывать.
5. B1 `therapy_ines_future_plan` - обсудить план изменений.
6. B1 `therapy_ines_old_photo` - рассказать, что чувствуешь из-за старого фото.

### Profesor Martín Sol, преподаватель, B1-B2

История: преподаватель культурологии собирает городские истории о площади, рынке и старом театре.

Квесты:

1. B1 `university_martin_join_seminar` - представиться на семинаре.
2. B1 `university_martin_explain_topic` - выбрать тему мини-доклада.
3. B1 `university_martin_interview_neighbor` - подготовить вопросы для интервью.
4. B2 `university_martin_compare_versions` - сравнить две версии легенды.
5. B2 `university_martin_defend_argument` - защитить аргумент.
6. B2 `university_martin_public_talk` - выступить с выводом.

### Doña Pilar, аптекарь, A1-A2

История: пожилая аптекарь с острым языком и добрым сердцем. Знает всё о районе, но выдаёт секреты только в обмен на вежливость.

Квесты:

1. A1 `pharmacy_pilar_cough` - попросить что-то от кашля.
2. A1 `pharmacy_pilar_price` - спросить цену и способ оплаты.
3. A2 `pharmacy_pilar_symptoms` - описать симптомы.
4. A2 `pharmacy_pilar_for_friend` - купить лекарство для друга.
5. A2 `pharmacy_pilar_warning` - понять предупреждение о дозировке.
6. A2 `pharmacy_pilar_gossip` - расспросить о странном посетителе.

### Dr. Mateo Cruz, врач, A2-B1

История: врач при маленькой клинике возле рынка. Очень практичный, но замечает, что многие пациенты видели один и тот же странный сон.

Квесты:

1. A2 `clinic_mateo_appointment` - записаться на приём.
2. A2 `clinic_mateo_pain` - описать боль.
3. A2 `clinic_mateo_habits` - ответить про сон, еду, работу.
4. B1 `clinic_mateo_when_started` - рассказать историю болезни.
5. B1 `clinic_mateo_follow_up` - обсудить, помогли ли советы.
6. B1 `clinic_mateo_shared_dream` - пересказать сон и сравнить с другими.

### El Hombre del Paraguas, городской чудак, A1-B1

История: человек с зонтом появляется даже в солнечную погоду и говорит, что "дождь бывает не только с неба". Постепенно это становится историей о потерянной любви и памяти.

Квесты:

1. A1 `umbrella_man_weather` - поговорить о погоде.
2. A1 `umbrella_man_direction` - спросить дорогу.
3. A2 `umbrella_man_warning` - понять простое предупреждение.
4. A2 `umbrella_man_lost_letter` - описать найденное письмо.
5. B1 `umbrella_man_old_promise` - обсудить обещание из прошлого.
6. B1 `umbrella_man_sunny_day` - помочь ему принять хороший день.

### Carmen Sombra, прохожая-сыщица, B1-B2

Рефрен: тень классического детектива в плаще и шляпе, без прямых имён.

История: Кармен расследует не преступление, а исчезновение городского мурала.

Квесты:

1. B1 `detective_carmen_first_clue` - описать, что видел.
2. B1 `detective_carmen_wall` - описать рисунок на стене.
3. B1 `detective_carmen_alibi` - объяснить, где был человек.
4. B2 `detective_carmen_symbol` - обсудить значение символа.
5. B2 `detective_carmen_false_lead` - признать ошибочную версию.
6. B2 `detective_carmen_mural_returns` - понять, почему мурал исчез.

### Rafa, водитель автобуса, A2-B1

История: болтливый водитель знает город по остановкам. Каждая остановка - маленькая история.

Квесты:

1. A2 `bus_rafa_ticket` - купить билет.
2. A2 `bus_rafa_next_stop` - спросить о следующей остановке.
3. A2 `bus_rafa_wrong_bus` - объяснить, что сел не туда.
4. B1 `bus_rafa_passenger_story` - пересказать историю пассажира.
5. B1 `bus_rafa_old_route` - узнать о старом маршруте.
6. B1 `bus_rafa_last_stop` - решить, стоит ли ехать до последней остановки.

### Lucía Canta, уличная певица, A1-B1

История: поёт на площади и собирает слова для песни о городе. Игрок помогает ей выбирать слова и истории.

Квесты:

1. A1 `singer_lucia_greeting` - поздороваться и похвалить песню.
2. A1 `singer_lucia_request` - попросить песню.
3. A2 `singer_lucia_describe_song` - описать настроение.
4. A2 `singer_lucia_missing_word` - предложить слово для строки.
5. B1 `singer_lucia_memory` - рассказать личное воспоминание для песни.
6. B1 `singer_lucia_new_song` - обсудить готовую песню.

### Bibliotecaria Vega, библиотекарь, B1-C1

История: хранит архив писем, которые никто не отправил. Это хорошая цепочка для пересказа, косвенной речи, предположений и нюансов.

Квесты:

1. B1 `library_vega_card` - оформить карточку.
2. B1 `library_vega_find_letter` - описать письмо, которое ищешь.
3. B1 `library_vega_summarize_letter` - пересказать содержание.
4. B2 `library_vega_infer_author` - предположить, кто автор.
5. B2 `library_vega_ethics` - обсудить, можно ли читать чужие письма.
6. C1 `library_vega_unsent_truth` - обсудить правду, которую человек не решился сказать.

### Don Álvaro, бывший "герой площади", B2-C1

Рефрен: старый авантюрист или рыцарь без прямых отсылок.

История: он считает себя защитником площади, но его настоящая битва - признать возраст, ошибки и одиночество.

Квесты:

1. B2 `old_hero_plaza_task` - помочь с "важным поручением".
2. B2 `old_hero_exaggerated_story` - отделить факты от преувеличений.
3. B2 `old_hero_public_argument` - урегулировать спор.
4. C1 `old_hero_regret` - обсудить сожаление.
5. C1 `old_hero_legacy` - поговорить о наследии.
6. C1 `old_hero_final_walk` - провести его по площади и подвести итог.

## Приоритет для ближайшей реализации

Рекомендуемый порядок:

1. A1-A2: `Nora Finch` / `Doña Pilar` и `Dr. Elias Stone` / `Dr. Mateo Cruz` - медицина, симптомы, инструкции.
2. A2-B1: психолог - рассказ о себе, эмоции, планы, привычки.
3. B1: водитель, музыкант, городской чудак - связные бытовые истории.
4. B2-C1: профессор, библиотекарь, бывший герой - аргументация, пересказ, абстрактные темы.

При разворачивании конкретного NPC в миграцию нужно подготовить:

- `scenario.code`;
- `npc_code`;
- `prerequisite_code`;
- `cefr_level`;
- `place_type`;
- `title`;
- `npc_name`;
- `npc_persona`;
- `scene_setup`;
- `is_quest`;
- `max_turns`;
- список `conversation_tasks` с `code`, `title`, `is_required`, `sort_order`, `completion_criteria`.
