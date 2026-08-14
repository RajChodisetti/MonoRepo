DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM outreach_email_credentials) THEN
    RAISE EXCEPTION 'refusing to remove migration 52 while encrypted outreach email credentials exist';
  END IF;
END $$;

DROP TABLE outreach_email_credentials;
