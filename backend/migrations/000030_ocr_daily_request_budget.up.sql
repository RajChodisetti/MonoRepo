-- Durable global daily budget for outbound menu OCR provider requests.

CREATE TABLE IF NOT EXISTS ocr_daily_request_usage (
  usage_date date NOT NULL,
  budget_key text NOT NULL,
  requests_used integer NOT NULL DEFAULT 0,
  daily_limit integer NOT NULL,
  last_requested_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (usage_date, budget_key),
  CONSTRAINT ocr_daily_request_usage_budget_key_check
    CHECK (length(trim(budget_key)) BETWEEN 1 AND 100),
  CONSTRAINT ocr_daily_request_usage_requests_check
    CHECK (requests_used >= 0),
  CONSTRAINT ocr_daily_request_usage_limit_check
    CHECK (daily_limit BETWEEN 1 AND 200),
  CONSTRAINT ocr_daily_request_usage_within_limit_check
    CHECK (requests_used <= daily_limit)
);

CREATE INDEX IF NOT EXISTS idx_ocr_daily_request_usage_recent
  ON ocr_daily_request_usage (usage_date DESC, budget_key);
