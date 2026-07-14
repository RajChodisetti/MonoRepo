ALTER TABLE scrape_jobs
  DROP CONSTRAINT IF EXISTS scrape_jobs_current_cell_fk;

DROP TABLE IF EXISTS scrape_job_candidates;
DROP TABLE IF EXISTS scrape_job_cells;
DROP TABLE IF EXISTS scrape_jobs;
