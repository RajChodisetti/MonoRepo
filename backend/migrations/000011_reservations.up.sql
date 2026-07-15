-- P1-021: table reservation requests (voice agent + demo site).

CREATE TABLE IF NOT EXISTS reservations (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  guest_name text NOT NULL,
  guest_phone text NOT NULL DEFAULT '',
  guest_email text NOT NULL DEFAULT '',
  party_size int NOT NULL CHECK (party_size >= 1 AND party_size <= 20),
  reservation_date date NOT NULL,
  reservation_time time NOT NULL,
  status text NOT NULL DEFAULT 'pending',
  source text NOT NULL DEFAULT 'voice_agent',
  notes text NOT NULL DEFAULT '',
  client_request_id text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_reservations_restaurant_date_status
  ON reservations (restaurant_id, reservation_date, status);

CREATE UNIQUE INDEX IF NOT EXISTS idx_reservations_client_request_id
  ON reservations (restaurant_id, client_request_id)
  WHERE client_request_id IS NOT NULL AND client_request_id <> '';
