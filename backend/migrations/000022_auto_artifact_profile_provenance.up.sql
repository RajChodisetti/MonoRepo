-- Bind automatically generated demo/campaign drafts to both the verified OCR
-- input and the exact restaurant identity/public payload used to create them.
-- Existing automatic artifacts have no trustworthy profile provenance, so
-- they are returned to draft and prepared again before another human review.

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

CREATE OR REPLACE FUNCTION lead_artifact_current_profile_fingerprint(p_restaurant_id uuid)
RETURNS text
LANGUAGE sql
STABLE
AS $$
  SELECT encode(
    digest(
      convert_to(
        jsonb_build_object(
          'restaurant_id', r.id::text,
          'restaurant_name', r.name,
          'recipient_email', lower(trim(COALESCE(r.email, ''))),
          'public_payload', lead_artifact_current_public_payload(r.id)
        )::text,
        'UTF8'
      ),
      'sha256'
    ),
    'hex'
  )
  FROM restaurants r
  WHERE r.id = p_restaurant_id
$$;

ALTER TABLE demo_sites
  ADD COLUMN IF NOT EXISTS source_profile_fingerprint text NOT NULL DEFAULT '';

ALTER TABLE demo_sites
  DROP CONSTRAINT IF EXISTS demo_sites_auto_profile_fingerprint_check;

ALTER TABLE demo_sites
  ADD CONSTRAINT demo_sites_auto_profile_fingerprint_check
  CHECK (
    source_profile_fingerprint = ''
    OR source_profile_fingerprint ~ '^[0-9a-f]{64}$'
  );

ALTER TABLE email_campaigns
  ADD COLUMN IF NOT EXISTS source_profile_fingerprint text NOT NULL DEFAULT '';

ALTER TABLE email_campaigns
  DROP CONSTRAINT IF EXISTS email_campaigns_auto_profile_fingerprint_check;

ALTER TABLE email_campaigns
  ADD CONSTRAINT email_campaigns_auto_profile_fingerprint_check
  CHECK (
    source_profile_fingerprint = ''
    OR source_profile_fingerprint ~ '^[0-9a-f]{64}$'
  );

-- A pre-000022 automatic artifact cannot prove which identity/profile/menu
-- payload an administrator reviewed. Fail closed and regenerate the draft.
UPDATE email_campaigns
SET status = 'draft',
    approved_at = NULL,
    approved_by = NULL,
    updated_at = now()
WHERE auto_generated = true
  AND source_profile_fingerprint = ''
  AND status = 'approved';

UPDATE demo_sites
SET status = 'draft',
    published_at = NULL,
    published_by = NULL,
    updated_at = now()
WHERE auto_generated = true
  AND source_profile_fingerprint = ''
  AND status = 'published';

INSERT INTO job_runs (
  job_type, status, payload, idempotency_key, max_attempts
)
SELECT
  'lead.prepare',
  'queued',
  jsonb_build_object('restaurant_id', rp.restaurant_id::text),
  'lead.prepare:profile-provenance:' || rp.restaurant_id::text || ':' ||
    rp.ocr_input_fingerprint || ':' ||
    lead_artifact_current_profile_fingerprint(rp.restaurant_id),
  3
FROM restaurant_profiles rp
WHERE rp.ocr_status = 'verified'
  AND EXISTS (
    SELECT 1
    FROM email_campaigns c
    WHERE c.restaurant_id = rp.restaurant_id
      AND c.auto_generated = true
      AND c.status = 'draft'
      AND c.source_profile_fingerprint = ''
  )
  AND lead_artifact_current_profile_fingerprint(rp.restaurant_id) IS NOT NULL
ON CONFLICT (idempotency_key) WHERE idempotency_key IS NOT NULL
DO UPDATE SET
  status = 'queued',
  payload = EXCLUDED.payload,
  attempts = 0,
  max_attempts = EXCLUDED.max_attempts,
  last_error = NULL,
  available_at = now(),
  locked_at = NULL,
  locked_by = NULL,
  lease_expires_at = NULL,
  updated_at = now()
WHERE job_runs.job_type = 'lead.prepare'
  AND job_runs.status IN ('completed', 'failed');
