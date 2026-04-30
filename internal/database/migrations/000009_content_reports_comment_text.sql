ALTER TABLE content_reports
ADD COLUMN comment_text TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_content_reports_source_status_created
ON content_reports(source_type, status, created_at DESC);
