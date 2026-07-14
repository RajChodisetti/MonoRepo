-- Keep provider-capped leaves visibly incomplete instead of recording a false
-- completed city pass when the configured grid depth is exhausted.

ALTER TABLE scrape_jobs
  DROP CONSTRAINT IF EXISTS scrape_jobs_waiting_reason_check;

ALTER TABLE scrape_jobs
  ADD CONSTRAINT scrape_jobs_waiting_reason_check
  CHECK (
    waiting_reason IS NULL OR waiting_reason IN (
      'request_limit', 'provider_error', 'coverage_incomplete', 'revisit'
    )
  );

ALTER TABLE scrape_job_cells
  DROP CONSTRAINT IF EXISTS scrape_job_cells_status_check;

ALTER TABLE scrape_job_cells
  ADD CONSTRAINT scrape_job_cells_status_check
  CHECK (
    status IN (
      'pending', 'running', 'subdivided', 'completed', 'coverage_incomplete', 'failed'
    )
  );
