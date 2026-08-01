ALTER TABLE menu_images ADD COLUMN IF NOT EXISTS hidden_at timestamptz;
ALTER TABLE menu_images ADD COLUMN IF NOT EXISTS hidden_by uuid;
ALTER TABLE gallery_images ADD COLUMN IF NOT EXISTS hidden_at timestamptz;
ALTER TABLE gallery_images ADD COLUMN IF NOT EXISTS hidden_by uuid;

CREATE INDEX IF NOT EXISTS idx_menu_images_visible
  ON menu_images (restaurant_id) WHERE hidden_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_gallery_images_visible
  ON gallery_images (restaurant_id) WHERE hidden_at IS NULL;
