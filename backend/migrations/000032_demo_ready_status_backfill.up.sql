-- Leads that already have generated demo artifacts should appear under the
-- demo_ready lifecycle filter. Preserve later lifecycle states such as emailed,
-- interested, client, lost, and archived.
UPDATE restaurants r
SET status = 'demo_ready',
    updated_at = now()
WHERE r.status = 'lead'
  AND EXISTS (
    SELECT 1
    FROM demo_sites d
    WHERE d.restaurant_id = r.id
      AND d.status IN ('draft', 'published')
  );
