-- P1-011: restaurant profiles, menus, menu items, and scraped review storage.

CREATE TABLE IF NOT EXISTS restaurant_profiles (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL UNIQUE REFERENCES restaurants(id) ON DELETE CASCADE,
  description text NOT NULL DEFAULT '',
  opening_hours jsonb NOT NULL DEFAULT '{}',
  phone text NOT NULL DEFAULT '',
  website text NOT NULL DEFAULT '',
  address text NOT NULL DEFAULT '',
  city text NOT NULL DEFAULT '',
  state text NOT NULL DEFAULT '',
  country text NOT NULL DEFAULT '',
  latitude double precision,
  longitude double precision,
  google_place_id text,
  google_data_id text,
  rating numeric(3,1),
  reviews_count int,
  price_level text NOT NULL DEFAULT '',
  cuisines jsonb NOT NULL DEFAULT '[]',
  owners jsonb NOT NULL DEFAULT '[]',
  images jsonb NOT NULL DEFAULT '{}',
  apollo_lead jsonb NOT NULL DEFAULT '{}',
  scrape_status text NOT NULL DEFAULT 'unknown',
  scrape_errors jsonb NOT NULL DEFAULT '[]',
  dietary_options jsonb NOT NULL DEFAULT '[]',
  parking_info text NOT NULL DEFAULT '',
  reservation_policy text NOT NULL DEFAULT '',
  brand_tone text NOT NULL DEFAULT '',
  raw_public_data jsonb NOT NULL DEFAULT '{}',
  review_status text NOT NULL DEFAULT 'draft',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_restaurant_profiles_google_place_id
  ON restaurant_profiles (google_place_id)
  WHERE google_place_id IS NOT NULL AND google_place_id <> '';

CREATE INDEX IF NOT EXISTS idx_restaurant_profiles_city
  ON restaurant_profiles (city);

CREATE TABLE IF NOT EXISTS menus (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_menus_restaurant_id
  ON menus (restaurant_id);

CREATE TABLE IF NOT EXISTS menu_items (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  menu_id uuid NOT NULL REFERENCES menus(id) ON DELETE CASCADE,
  name text NOT NULL,
  description text NOT NULL DEFAULT '',
  price numeric(10,2),
  price_text text NOT NULL DEFAULT '',
  category text NOT NULL DEFAULT '',
  image_url text NOT NULL DEFAULT '',
  images jsonb NOT NULL DEFAULT '[]',
  is_available boolean NOT NULL DEFAULT true,
  sort_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_menu_items_menu_id
  ON menu_items (menu_id);

CREATE TABLE IF NOT EXISTS restaurant_reviews (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  reviewer text NOT NULL DEFAULT '',
  review_text text NOT NULL DEFAULT '',
  stars numeric(2,1),
  review_date text NOT NULL DEFAULT '',
  images jsonb NOT NULL DEFAULT '[]',
  source text NOT NULL DEFAULT '',
  sort_order int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_restaurant_reviews_restaurant_id
  ON restaurant_reviews (restaurant_id);

CREATE TABLE IF NOT EXISTS restaurant_data_imports (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  source_file text NOT NULL,
  meta jsonb NOT NULL DEFAULT '{}',
  restaurants_imported int NOT NULL DEFAULT 0,
  restaurants_skipped int NOT NULL DEFAULT 0,
  created_at timestamptz NOT NULL DEFAULT now()
);
