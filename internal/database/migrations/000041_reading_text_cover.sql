-- Cover image paths for reading texts (synced from bundle JSON top-level fields).
ALTER TABLE reading_texts ADD COLUMN IF NOT EXISTS cover_thumb_rel_path TEXT;
ALTER TABLE reading_texts ADD COLUMN IF NOT EXISTS cover_hero_rel_path TEXT;
ALTER TABLE reading_texts ADD COLUMN IF NOT EXISTS cover_image_prompt TEXT;
