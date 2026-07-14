UPDATE scrape_jobs
SET waiting_reason = 'provider_error'
WHERE waiting_reason = 'coverage_incomplete';

UPDATE scrape_job_cells
SET status = 'pending', saturated = false, last_error = NULL,
    page_token = NULL, page_number = 0, results_seen_cycle = 0,
    started_at = NULL, completed_at = NULL, updated_at = now()
WHERE status = 'coverage_incomplete';

ALTER TABLE scrape_job_cells
  DROP CONSTRAINT IF EXISTS scrape_job_cells_status_check;

ALTER TABLE scrape_job_cells
  ADD CONSTRAINT scrape_job_cells_status_check
  CHECK (
    status IN ('pending', 'running', 'subdivided', 'completed', 'failed')
  );

ALTER TABLE scrape_jobs
  DROP CONSTRAINT IF EXISTS scrape_jobs_waiting_reason_check;

ALTER TABLE scrape_jobs
  ADD CONSTRAINT scrape_jobs_waiting_reason_check
  CHECK (
    waiting_reason IS NULL OR waiting_reason IN (
      'request_limit', 'provider_error', 'revisit'
    )
  );
