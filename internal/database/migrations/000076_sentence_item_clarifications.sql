-- Short, learner-visible context used only when the Russian source sentence does
-- not by itself determine an article, pronoun, or other required target form.
ALTER TABLE sentence_items
    ADD COLUMN IF NOT EXISTS clarification_ru TEXT NOT NULL DEFAULT '';
