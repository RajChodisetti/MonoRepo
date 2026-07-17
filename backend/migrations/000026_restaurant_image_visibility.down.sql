DROP INDEX IF EXISTS idx_gallery_images_visible;
DROP INDEX IF EXISTS idx_menu_images_visible;

ALTER TABLE gallery_images DROP COLUMN IF EXISTS hidden_by;
ALTER TABLE gallery_images DROP COLUMN IF EXISTS hidden_at;
ALTER TABLE menu_images DROP COLUMN IF EXISTS hidden_by;
ALTER TABLE menu_images DROP COLUMN IF EXISTS hidden_at;
