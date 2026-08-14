# Database-managed outreach email accounts

## Context

Outreach Gmail accounts were supplied only through protected environment JSON.
Adding another mailbox therefore required editing deployment configuration and
restarting the API and worker. Operators need a secure admin path to add future
mailboxes while retaining existing environment accounts, avoiding duplicate
sending/polling when the same identity exists in both sources, and never exposing
OAuth material through browser reads.

## Decision

- Keep environment accounts supported and read-only. They take precedence when
  either their stable account key or normalized mailbox collides with a database
  row.
- Store UI-managed mailbox identity in PostgreSQL and its OAuth client ID,
  client secret, and refresh token as one AES-256-GCM ciphertext. The application
  key is supplied separately through `OUTREACH_CREDENTIAL_ENCRYPTION_KEY` as
  standard base64 encoding of exactly 32 bytes.
- Bind ciphertext authentication to the immutable account key and mailbox.
  Credentials are write-only at the HTTP boundary and never returned or logged.
- Allow internal administrators to add, enable, disable, and replace the
  credentials for database accounts. Preserve disabled rows and their message,
  quota, and inbox-sync history.
- Reload the union of environment and enabled database accounts at sending,
  health synchronization, and each inbox polling cycle. Database additions do
  not require a process restart and do not enable the bulk email job.
- Require each mailbox token to authorize `gmail.send` and `gmail.readonly`,
  because all effective sending accounts also participate in the unified inbox.

## Options

- Environment-only configuration: operationally safe but requires a deployment
  for every account change.
- Store plaintext OAuth fields: simpler but unnecessarily exposes durable
  credentials to database readers and backups.
- Replace environment configuration with database rows: rejected because it
  creates a risky cutover and removes the protected rollback source.

## Consequences

The application needs a protected credential-encryption key in every environment
that allows UI-managed accounts. Loss of that key makes database credentials
unusable until replaced, but environment accounts continue to work. Key rotation
requires a future re-encryption workflow. Runtime reloads add small database and
provider-construction overhead to outreach operations in exchange for immediate,
auditable changes.

## Rollback/Revisit Trigger

Disable database accounts to return immediately to environment-only behavior.
Migration `000052` can be rolled back only when the credential table is empty.
Revisit with envelope/KMS encryption if account volume grows materially or key
rotation becomes a compliance requirement.
