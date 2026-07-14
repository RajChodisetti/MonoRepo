CREATE TABLE IF NOT EXISTS scrape_jobs (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  city text NOT NULL,
  city_key text NOT NULL,
  niche text NOT NULL DEFAULT 'restaurant',
  status text NOT NULL DEFAULT 'queued',
  cycle_number integer NOT NULL DEFAULT 1,
  max_requests_per_window integer NOT NULL DEFAULT 500,
  requests_used_window integer NOT NULL DEFAULT 0,
  requests_used_total bigint NOT NULL DEFAULT 0,
  window_started_at timestamptz,
  resume_at timestamptz,
  waiting_reason text,
  last_cycle_completed_at timestamptz,
  current_cell_id uuid,
  locked_by text,
  locked_at timestamptz,
  lease_expires_at timestamptz,
  last_error text,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT scrape_jobs_status_check CHECK (
    status IN ('queued', 'running', 'waiting', 'completed', 'failed', 'cancelled')
  ),
  CONSTRAINT scrape_jobs_waiting_reason_check CHECK (
    waiting_reason IS NULL OR waiting_reason IN ('request_limit', 'provider_error', 'revisit')
  ),
  CONSTRAINT scrape_jobs_cycle_number_check CHECK (cycle_number >= 1),
  CONSTRAINT scrape_jobs_request_limit_check CHECK (
    max_requests_per_window BETWEEN 1 AND 500
  ),
  CONSTRAINT scrape_jobs_request_count_check CHECK (
    requests_used_window >= 0
    AND requests_used_window <= max_requests_per_window
    AND requests_used_total >= requests_used_window
  )
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_scrape_jobs_active_city_niche
  ON scrape_jobs (city_key, niche)
  WHERE status IN ('queued', 'running', 'waiting');

CREATE INDEX IF NOT EXISTS idx_scrape_jobs_due
  ON scrape_jobs (status, resume_at, created_at);

CREATE INDEX IF NOT EXISTS idx_scrape_jobs_lease
  ON scrape_jobs (lease_expires_at)
  WHERE status = 'running';

CREATE TABLE IF NOT EXISTS scrape_job_cells (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scrape_job_id uuid NOT NULL REFERENCES scrape_jobs(id) ON DELETE CASCADE,
  parent_cell_id uuid,
  cell_key text NOT NULL,
  depth integer NOT NULL DEFAULT 0,
  low_lat double precision NOT NULL,
  low_lng double precision NOT NULL,
  high_lat double precision NOT NULL,
  high_lng double precision NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  cycle_number integer NOT NULL DEFAULT 1,
  page_token text,
  page_number integer NOT NULL DEFAULT 0,
  results_seen_cycle integer NOT NULL DEFAULT 0,
  results_seen_total bigint NOT NULL DEFAULT 0,
  saturated boolean NOT NULL DEFAULT false,
  last_error text,
  started_at timestamptz,
  completed_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT scrape_job_cells_status_check CHECK (
    status IN ('pending', 'running', 'subdivided', 'completed', 'failed')
  ),
  CONSTRAINT scrape_job_cells_bounds_check CHECK (
    low_lat >= -90 AND high_lat <= 90 AND low_lat < high_lat
    AND low_lng >= -180 AND high_lng <= 180 AND low_lng < high_lng
  ),
  CONSTRAINT scrape_job_cells_depth_check CHECK (depth >= 0),
  CONSTRAINT scrape_job_cells_page_check CHECK (
    page_number >= 0 AND results_seen_cycle >= 0 AND results_seen_total >= results_seen_cycle
  ),
  CONSTRAINT scrape_job_cells_cycle_check CHECK (cycle_number >= 1),
  CONSTRAINT scrape_job_cells_job_key_unique UNIQUE (scrape_job_id, cell_key),
  CONSTRAINT scrape_job_cells_job_id_id_unique UNIQUE (scrape_job_id, id),
  CONSTRAINT scrape_job_cells_parent_same_job_fk
    FOREIGN KEY (scrape_job_id, parent_cell_id)
    REFERENCES scrape_job_cells (scrape_job_id, id)
    ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_scrape_job_cells_next
  ON scrape_job_cells (scrape_job_id, status, depth, cell_key);

CREATE INDEX IF NOT EXISTS idx_scrape_job_cells_parent
  ON scrape_job_cells (parent_cell_id);

ALTER TABLE scrape_jobs
  DROP CONSTRAINT IF EXISTS scrape_jobs_current_cell_fk;

ALTER TABLE scrape_jobs
  ADD CONSTRAINT scrape_jobs_current_cell_fk
  FOREIGN KEY (id, current_cell_id)
  REFERENCES scrape_job_cells (scrape_job_id, id)
  ON DELETE SET NULL (current_cell_id);

CREATE TABLE IF NOT EXISTS scrape_job_candidates (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  scrape_job_id uuid NOT NULL REFERENCES scrape_jobs(id) ON DELETE CASCADE,
  first_cell_id uuid,
  google_place_id text NOT NULL,
  status text NOT NULL DEFAULT 'discovered',
  first_seen_cycle integer NOT NULL,
  last_seen_cycle integer NOT NULL,
  discovery_data jsonb NOT NULL DEFAULT '{}',
  scrape_record jsonb NOT NULL DEFAULT '{}',
  restaurant_id uuid REFERENCES restaurants(id) ON DELETE SET NULL,
  attempts integer NOT NULL DEFAULT 0,
  last_error text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT scrape_job_candidates_status_check CHECK (
    status IN ('discovered', 'details_ready', 'enriched', 'imported', 'duplicate', 'failed')
  ),
  CONSTRAINT scrape_job_candidates_cycle_check CHECK (
    first_seen_cycle >= 1 AND last_seen_cycle >= first_seen_cycle
  ),
  CONSTRAINT scrape_job_candidates_attempts_check CHECK (attempts >= 0),
  CONSTRAINT scrape_job_candidates_place_unique UNIQUE (scrape_job_id, google_place_id),
  CONSTRAINT scrape_job_candidates_first_cell_same_job_fk
    FOREIGN KEY (scrape_job_id, first_cell_id)
    REFERENCES scrape_job_cells (scrape_job_id, id)
    ON DELETE SET NULL (first_cell_id)
);

CREATE INDEX IF NOT EXISTS idx_scrape_job_candidates_next
  ON scrape_job_candidates (scrape_job_id, status, created_at);

CREATE INDEX IF NOT EXISTS idx_scrape_job_candidates_place
  ON scrape_job_candidates (google_place_id);
