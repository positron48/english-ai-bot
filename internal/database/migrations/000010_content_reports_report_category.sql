ALTER TABLE content_reports
ADD COLUMN IF NOT EXISTS report_category TEXT NOT NULL DEFAULT 'other';

CREATE INDEX IF NOT EXISTS idx_content_reports_source_category_status
ON content_reports(source_type, report_category, status);
