ALTER TABLE content_reports
    ADD COLUMN IF NOT EXISTS client_report_id TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_content_reports_client_report_id
    ON content_reports(user_id, client_report_id)
    WHERE client_report_id IS NOT NULL;
