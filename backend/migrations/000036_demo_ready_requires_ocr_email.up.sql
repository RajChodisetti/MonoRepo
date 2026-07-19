-- demo_ready means the lead can enter demo/outreach review: OCR verified and a
-- usable contact email is present. Draft demo artifacts alone are not enough.
UPDATE restaurants r
SET status = 'lead',
    updated_at = now()
WHERE r.status = 'demo_ready'
  AND (
    NULLIF(BTRIM(r.email), '') IS NULL
    OR NOT EXISTS (
      SELECT 1
      FROM restaurant_profiles rp
      WHERE rp.restaurant_id = r.id
        AND rp.ocr_status = 'verified'
    )
  );

UPDATE restaurants r
SET status = 'demo_ready',
    updated_at = now()
FROM restaurant_profiles rp
WHERE rp.restaurant_id = r.id
  AND r.status = 'lead'
  AND rp.ocr_status = 'verified'
  AND NULLIF(BTRIM(r.email), '') IS NOT NULL;
