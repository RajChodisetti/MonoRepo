-- A lead is OCR verified only after every scraped photo has a successful
-- structured vision result. Older "one successful photo is enough" rows must
-- return through OCR and human review before they can be published or sent.

UPDATE demo_sites d
SET status = 'draft',
    published_at = NULL,
    published_by = NULL,
    updated_at = now()
FROM restaurant_profiles rp
WHERE rp.restaurant_id = d.restaurant_id
  AND rp.ocr_status = 'verified'
  AND COALESCE(
        rp.raw_public_data #> '{menu_ocr,all_images_processed}',
        'false'::jsonb
      ) <> 'true'::jsonb
  AND d.status = 'published';

UPDATE email_campaigns c
SET status = 'draft',
    approved_at = NULL,
    approved_by = NULL,
    updated_at = now()
FROM restaurant_profiles rp
WHERE rp.restaurant_id = c.restaurant_id
  AND rp.ocr_status = 'verified'
  AND COALESCE(
        rp.raw_public_data #> '{menu_ocr,all_images_processed}',
        'false'::jsonb
      ) <> 'true'::jsonb
  AND c.status = 'approved';

UPDATE restaurant_profiles
SET ocr_status = 'pending',
    ocr_verified = false,
    ocr_verified_at = NULL,
    ocr_completed_at = NULL,
    ocr_claim_id = NULL,
    ocr_claim_fingerprint = NULL,
    ocr_verification_errors = COALESCE(
      ocr_verification_errors,
      '[]'::jsonb
    ) || '["full scraped-photo OCR verification required"]'::jsonb,
    review_status = 'draft',
    reviewed_at = NULL,
    reviewed_by = NULL,
    updated_at = now()
WHERE ocr_status = 'verified'
  AND COALESCE(
        raw_public_data #> '{menu_ocr,all_images_processed}',
        'false'::jsonb
      ) <> 'true'::jsonb;

ALTER TABLE restaurant_profiles
  DROP CONSTRAINT IF EXISTS restaurant_profiles_ocr_full_photo_check;

ALTER TABLE restaurant_profiles
  ADD CONSTRAINT restaurant_profiles_ocr_full_photo_check
  CHECK (
    ocr_status <> 'verified'
    OR COALESCE(
         raw_public_data #> '{menu_ocr,all_images_processed}',
         'false'::jsonb
       ) = 'true'::jsonb
  );
