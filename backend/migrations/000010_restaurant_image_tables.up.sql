-- OCR-classified restaurant images: printed menus vs gallery (food, interior, etc.)

CREATE TABLE IF NOT EXISTS menu_images (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  url text NOT NULL,
  thumbnail_url text NOT NULL DEFAULT '',
  image_type text NOT NULL DEFAULT 'menu_document',
  confidence numeric(4,3),
  title text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT 'menu_ocr',
  sort_order int NOT NULL DEFAULT 0,
  metadata jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_menu_images_restaurant_url
  ON menu_images (restaurant_id, url);

CREATE INDEX IF NOT EXISTS idx_menu_images_restaurant_id
  ON menu_images (restaurant_id);

CREATE TABLE IF NOT EXISTS gallery_images (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  restaurant_id uuid NOT NULL REFERENCES restaurants(id) ON DELETE CASCADE,
  url text NOT NULL,
  thumbnail_url text NOT NULL DEFAULT '',
  image_type text NOT NULL DEFAULT 'other',
  confidence numeric(4,3),
  title text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT 'menu_ocr',
  sort_order int NOT NULL DEFAULT 0,
  metadata jsonb NOT NULL DEFAULT '{}',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_gallery_images_restaurant_url
  ON gallery_images (restaurant_id, url);

CREATE INDEX IF NOT EXISTS idx_gallery_images_restaurant_id
  ON gallery_images (restaurant_id);
