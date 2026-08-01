UPDATE restaurant_profiles
SET
  ocr_status = 'pending',
  ocr_verified = false,
  ocr_verified_at = NULL,
  ocr_started_at = NULL,
  ocr_completed_at = NULL,
  ocr_attempts = 0,
  ocr_claim_id = NULL,
  ocr_claim_fingerprint = NULL,
  ocr_verification_errors = COALESCE(ocr_verification_errors, '[]'::jsonb)
    || '["Reset fingerprintless Google Places OCR classifications for media reprocessing"]'::jsonb,
  raw_public_data = jsonb_set(
    COALESCE(raw_public_data, '{}'::jsonb),
    '{menu_ocr}',
    (
      (COALESCE(raw_public_data->'menu_ocr', '{}'::jsonb) - 'classifications')
      || jsonb_build_object(
        'reset_reason', 'fingerprintless_google_places_classifications',
        'reset_at', to_jsonb(now())
      )
    ),
    true
  ),
  updated_at = now()
WHERE raw_public_data IS NOT NULL
  AND jsonb_typeof(raw_public_data->'menu_ocr'->'classifications') = 'array'
  AND EXISTS (
    SELECT 1
    FROM jsonb_array_elements(raw_public_data->'menu_ocr'->'classifications') AS classification(value)
    WHERE classification.value->>'source' = 'google_places_photo'
      AND COALESCE(classification.value->>'source_fingerprint', '') = ''
  );
