CREATE INDEX IF NOT EXISTS idx_content_reports_status_source_id
ON content_reports(status, source_type, id DESC);

CREATE INDEX IF NOT EXISTS idx_content_reports_status_source_chapter
ON content_reports(status, source_type, grammar_chapter_id);

CREATE INDEX IF NOT EXISTS idx_content_reports_status_source_theory
ON content_reports(status, source_type, theory_block_id);
