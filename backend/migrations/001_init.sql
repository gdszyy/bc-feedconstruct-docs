-- BDD scaffold migration — final DDL is committed after the storage
-- BDD empty tests are confirmed by the user.
--
-- Authoritative schema spec: docs/08_backend_railway/02_postgres_schema.md

-- placeholder: enables uuid generation
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- raw_messages will be added in 001 (final pass), kept empty here on purpose
-- so that running this scaffold migration is a no-op beyond the extension.
