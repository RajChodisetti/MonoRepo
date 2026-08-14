-- Internal admins may add Gmail outreach accounts without changing deployment
-- configuration. OAuth material is encrypted by the application before it is
-- stored; mailbox identity remains plaintext for deterministic deduplication.

CREATE TABLE outreach_email_credentials (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  account_key text NOT NULL UNIQUE,
  mailbox_email text NOT NULL UNIQUE,
  from_email text NOT NULL,
  credential_ciphertext bytea NOT NULL,
  encryption_version smallint NOT NULL DEFAULT 1,
  enabled boolean NOT NULL DEFAULT true,
  created_by uuid REFERENCES users(id) ON DELETE SET NULL,
  updated_by uuid REFERENCES users(id) ON DELETE SET NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CONSTRAINT outreach_email_credentials_key_check
    CHECK (account_key ~ '^[a-z0-9][a-z0-9_-]{1,63}$'),
  CONSTRAINT outreach_email_credentials_mailbox_check
    CHECK (
      mailbox_email = lower(btrim(mailbox_email))
      AND length(mailbox_email) BETWEEN 3 AND 320
      AND position('@' IN mailbox_email) > 1
    ),
  CONSTRAINT outreach_email_credentials_from_check
    CHECK (
      from_email = lower(btrim(from_email))
      AND length(from_email) BETWEEN 3 AND 320
      AND position('@' IN from_email) > 1
    ),
  CONSTRAINT outreach_email_credentials_ciphertext_check
    CHECK (octet_length(credential_ciphertext) >= 32),
  CONSTRAINT outreach_email_credentials_encryption_version_check
    CHECK (encryption_version = 1)
);

CREATE INDEX outreach_email_credentials_enabled
  ON outreach_email_credentials (enabled, created_at, account_key);
