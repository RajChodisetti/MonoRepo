CREATE OR REPLACE FUNCTION lead_artifact_current_public_payload(p_restaurant_id uuid)
RETURNS jsonb
LANGUAGE sql
STABLE
AS $$
  SELECT jsonb_build_object(
    'restaurant_name', r.name,
    'cuisine', COALESCE(rp.cuisines->>0, ''),
    'hero', COALESCE(
      NULLIF(rp.images->>'thumbnail', ''),
      (
        SELECT NULLIF(gi.url, '')
        FROM gallery_images gi
        WHERE gi.restaurant_id = r.id AND gi.url <> ''
        ORDER BY gi.sort_order ASC, gi.created_at ASC, gi.id ASC
        LIMIT 1
      ),
      ''
    ),
    'hours', COALESCE(rp.opening_hours, '{}'::jsonb),
    'address', COALESCE(rp.address, ''),
    'phone', COALESCE(rp.phone, ''),
    'menu_sections', COALESCE(
      (
        SELECT jsonb_agg(grouped.section ORDER BY grouped.section_name)
        FROM (
          SELECT
            COALESCE(NULLIF(mi.category, ''), 'Menu') AS section_name,
            jsonb_build_object(
              'name', COALESCE(NULLIF(mi.category, ''), 'Menu'),
              'items', jsonb_agg(
                jsonb_build_object(
                  'name', mi.name,
                  'description', mi.description,
                  'price', COALESCE(
                    NULLIF(mi.price_text, ''),
                    CASE
                      WHEN mi.price IS NULL THEN ''
                      ELSE '$' || trim(to_char(mi.price, 'FM999999990.00'))
                    END
                  ),
                  'image_url', mi.image_url
                )
                ORDER BY mi.sort_order ASC, mi.name ASC, mi.id ASC
              )
            ) AS section
          FROM menu_items mi
          JOIN menus m ON m.id = mi.menu_id
          WHERE m.restaurant_id = r.id AND m.name = 'Imported Menu'
          GROUP BY COALESCE(NULLIF(mi.category, ''), 'Menu')
        ) grouped
      ),
      '[]'::jsonb
    ),
    'reservation_cta', 'Book a table',
    'ai_receptionist_cta', 'Call our AI assistant'
  )
  FROM restaurants r
  JOIN restaurant_profiles rp ON rp.restaurant_id = r.id
  WHERE r.id = p_restaurant_id
$$;

DROP TABLE IF EXISTS restaurant_media_assets;
