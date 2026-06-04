ALTER TABLE learning_items
    DROP CONSTRAINT IF EXISTS learning_items_item_type_check;

ALTER TABLE learning_items
    ADD CONSTRAINT learning_items_item_type_check
    CHECK (item_type IN (
        'word',
        'grammar_chapter',
        'grammar_concept',
        'grammar_theory_block',
        'grammar_question',
        'reading_text',
        'reading_question',
        'speaking_task',
        'chat_correction',
        'pronunciation'
    ));
