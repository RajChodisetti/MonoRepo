-- Restore the previous demo_ready backfill rule used by migration 000032.
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
