-- Remove old multi-template outreach email content from unsent campaign rows.
-- Runtime preview/send also renders the current template, but rewriting stored
-- drafts prevents stale ghost links from reappearing in admin review screens.

WITH current_copy AS (
  SELECT
    c.id,
    COALESCE(NULLIF(TRIM(r.name), ''), 'your restaurant') AS restaurant_name,
    replace(
      replace(
        replace(COALESCE(NULLIF(TRIM(r.name), ''), 'your restaurant'), '&', '&amp;'),
        '<', '&lt;'
      ),
      '>', '&gt;'
    ) AS restaurant_name_html
  FROM email_campaigns c
  JOIN restaurants r ON r.id = c.restaurant_id
  WHERE c.campaign_type = 'outreach'
    AND c.status IN ('draft', 'approved')
    AND (
      c.body_html ILIKE '%TEMPLATE_1_URL%'
      OR c.body_html ILIKE '%TEMPLATE_2_URL%'
      OR c.body_html ILIKE '%TEMPLATE_3_URL%'
      OR c.body_html ILIKE '%Cinematic personalized website%'
      OR c.body_html ILIKE '%Aurora personalized website%'
      OR c.body_html ILIKE '%Elysian personalized website%'
      OR c.body_html ILIKE '%Open % demo%'
      OR c.body_html ILIKE '%Tuvi restaurant services presentation%'
      OR c.body_html ILIKE '%live website preview for %</a>%'
    )
)
UPDATE email_campaigns c
SET
  subject = 'A live demo for ' || current_copy.restaurant_name || ' — AI receptionist, website & more',
  body_html =
    '<!DOCTYPE html>' ||
    '<html lang="en"><head><meta charset="UTF-8" /><meta name="viewport" content="width=device-width, initial-scale=1.0" />' ||
    '<title>Your live restaurant demo</title></head>' ||
    '<body style="margin:0;padding:0;background:#f4f5f7;font-family:-apple-system,BlinkMacSystemFont,''Segoe UI'',Roboto,Helvetica,Arial,sans-serif;color:#111111;">' ||
    '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background:#f4f5f7;padding:28px 12px;"><tr><td align="center">' ||
    '<table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="max-width:560px;background:#ffffff;border:1px solid #e6e8ec;border-radius:16px;overflow:hidden;">' ||
    '<tr><td style="background:#111111;padding:0;"><table role="presentation" width="100%" cellspacing="0" cellpadding="0"><tr><td style="height:4px;background:#d4a853;font-size:0;line-height:0;">&nbsp;</td></tr><tr><td align="center" style="padding:22px 24px 18px;"><p style="margin:0;font-size:11px;letter-spacing:0.22em;font-weight:700;color:#d4a853;text-transform:uppercase;">Tuvi Solutions</p><p style="margin:8px 0 0;font-size:11px;letter-spacing:0.14em;font-weight:600;color:#9ca3af;text-transform:uppercase;">Restaurant growth platform</p></td></tr></table></td></tr>' ||
    '<tr><td style="padding:28px 24px 8px;"><p style="margin:0;font-size:15px;line-height:1.55;color:#5b6470;">Hi there — missed calls and outdated websites can make it harder for guests to request a table.</p><p style="margin:12px 0 0;font-size:15px;line-height:1.55;color:#374151;"><strong style="color:#111111;">Tuvi Solutions</strong> builds restaurant tools to help with that — and we''ve already created a live website preview for <strong style="color:#111111;">' || current_copy.restaurant_name_html || '</strong>, plus a reservation-request form and details on our AI receptionist.</p></td></tr>' ||
    '<tr><td style="padding:16px 24px 8px;"><p style="margin:0 0 10px;font-size:11px;letter-spacing:0.14em;font-weight:700;color:#6b7280;text-transform:uppercase;">Links for ' || current_copy.restaurant_name_html || '</p><table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="border:1px solid #e6e8ec;border-radius:12px;overflow:hidden;">' ||
    '<tr><td style="padding:14px 16px;"><a href="{{CLICK_URL}}" style="color:#111111;text-decoration:none;font-size:14px;font-weight:700;line-height:1.3;">Personalized demo websites →</a><p style="margin:4px 0 0;font-size:12px;line-height:1.45;color:#6b7280;">Open the live restaurant-specific website preview we prepared</p></td></tr>' ||
    '<tr><td style="padding:14px 16px;border-top:1px solid #e6e8ec;"><a href="https://tuvisolutions.com/services/restaurants" style="color:#111111;text-decoration:none;font-size:14px;font-weight:700;line-height:1.3;">Services catalog →</a><p style="margin:4px 0 0;font-size:12px;line-height:1.45;color:#6b7280;">Explore Tuvi''s restaurant services, including websites, AI receptionist, reservations, and growth tools</p></td></tr>' ||
    '</table></td></tr>' ||
    '<tr><td style="padding:20px 24px 28px;"><p style="margin:0;font-size:13px;line-height:1.55;color:#5b6470;">Reply anytime — happy to walk you through it in 10 minutes.</p><p style="margin:14px 0 0;font-size:12px;line-height:1.5;color:#9ca3af;">Prefer not to get these emails? <a href="{{UNSUBSCRIBE_URL}}" style="color:#6b7280;text-decoration:underline;">Unsubscribe</a>.</p><p style="margin:14px 0 0;font-size:13px;color:#111111;font-weight:600;">— Tuvi Solutions</p></td></tr>' ||
    '</table></td></tr></table></body></html>',
  body_text =
    'Hi there — missed calls and outdated websites can make it harder for guests to request a table.' || E'\n\n' ||
    'Tuvi Solutions builds restaurant tools to help with that — and we''ve already created a live website preview for ' || current_copy.restaurant_name || ', plus a reservation-request form and details on our AI receptionist.' || E'\n\n' ||
    'Links for ' || current_copy.restaurant_name || ':' || E'\n' ||
    '- Personalized demo websites: Open the live restaurant-specific website preview we prepared' || E'\n' ||
    '  {{CLICK_URL}}' || E'\n' ||
    '- Services catalog: Explore Tuvi''s restaurant services, including websites, AI receptionist, reservations, and growth tools' || E'\n' ||
    '  https://tuvisolutions.com/services/restaurants' || E'\n\n' ||
    'Reply anytime — happy to walk you through it in 10 minutes.' || E'\n\n' ||
    'Prefer not to get these emails? Unsubscribe: {{UNSUBSCRIBE_URL}}' || E'\n\n' ||
    '— Tuvi Solutions',
  updated_at = now()
FROM current_copy
WHERE c.id = current_copy.id;
