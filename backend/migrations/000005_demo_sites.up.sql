-- P1-009: public demo sites with token-gated access.

CREATE TABLE IF NOT EXISTS demo_sites (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  slug text NOT NULL UNIQUE,
  token_hash text NOT NULL,
  status text NOT NULL DEFAULT 'draft',
  public_payload jsonb NOT NULL DEFAULT '{}',
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_demo_sites_restaurant_id
  ON demo_sites (restaurant_id);
