CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    email_verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ
);

CREATE TABLE email_otps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    normalized_email TEXT NOT NULL,
    code_hash TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    consumed_at TIMESTAMPTZ,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE profiles (
    account_id UUID PRIMARY KEY REFERENCES accounts(id),
    handle TEXT NOT NULL UNIQUE,
    CHECK (handle ~ '^[a-z][a-z0-9_]{2,23}$'),
    display_name TEXT NOT NULL DEFAULT '',
    bio TEXT NOT NULL DEFAULT '',
    public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE agents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_account_id UUID NOT NULL UNIQUE REFERENCES accounts(id),
    provider_configured BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE rooms (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug TEXT NOT NULL UNIQUE,
    title TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE room_memberships (
    room_id UUID NOT NULL REFERENCES rooms(id),
    account_id UUID NOT NULL REFERENCES accounts(id),
    joined_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (room_id, account_id)
);

CREATE TABLE room_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id UUID NOT NULL REFERENCES rooms(id),
    author_account_id UUID NOT NULL REFERENCES accounts(id),
    content TEXT NOT NULL CHECK (char_length(content) BETWEEN 1 AND 4000),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE notes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_account_id UUID NOT NULL REFERENCES accounts(id),
    title TEXT NOT NULL DEFAULT '',
    body TEXT NOT NULL DEFAULT '',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE connectors (
    id TEXT NOT NULL,
    version TEXT NOT NULL,
    publisher_account_id UUID REFERENCES accounts(id),
    manifest JSONB NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('review','approved','rejected')),
    submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (id, version)
);

CREATE TABLE connector_installs (
    account_id UUID NOT NULL REFERENCES accounts(id),
    connector_id TEXT NOT NULL,
    connector_version TEXT NOT NULL,
    installed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, connector_id),
    FOREIGN KEY (connector_id, connector_version) REFERENCES connectors(id, version)
);

CREATE TABLE connector_grants (
    account_id UUID NOT NULL,
    connector_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (account_id, connector_id, operation),
    FOREIGN KEY (account_id, connector_id) REFERENCES connector_installs(account_id, connector_id) ON DELETE CASCADE
);

CREATE TABLE tool_audit (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    account_id UUID NOT NULL REFERENCES accounts(id),
    connector_id TEXT NOT NULL,
    operation TEXT NOT NULL,
    allowed BOOLEAN NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_account_id UUID NOT NULL REFERENCES accounts(id),
    reported_account_id UUID REFERENCES accounts(id),
    reported_message_id UUID REFERENCES room_messages(id),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (reported_account_id IS NOT NULL OR reported_message_id IS NOT NULL)
);

CREATE TABLE blocks (
    blocker_account_id UUID NOT NULL REFERENCES accounts(id),
    blocked_account_id UUID NOT NULL REFERENCES accounts(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (blocker_account_id, blocked_account_id),
    CHECK (blocker_account_id <> blocked_account_id)
);

CREATE INDEX idx_room_messages_room_created ON room_messages(room_id, created_at DESC);
CREATE INDEX idx_notes_account_updated ON notes(owner_account_id, updated_at DESC);
CREATE INDEX idx_tool_audit_account_occurred ON tool_audit(account_id, occurred_at DESC);
CREATE INDEX idx_email_otps_email_created ON email_otps(normalized_email, created_at DESC);
