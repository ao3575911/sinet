CREATE TABLE accounts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), email text NOT NULL UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE email_otps (
 email text PRIMARY KEY, nonce text NOT NULL, code_hash text NOT NULL,
 expires_at timestamptz NOT NULL, attempts integer NOT NULL DEFAULT 0,
 consumed boolean NOT NULL DEFAULT false
);
CREATE TABLE sessions (
 token_hash text PRIMARY KEY, account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
 expires_at timestamptz NOT NULL DEFAULT now() + interval '24 hours'
);
CREATE TABLE profiles (
 account_id uuid PRIMARY KEY REFERENCES accounts ON DELETE CASCADE,
 handle text UNIQUE CHECK (handle ~ '^[a-z][a-z0-9_]{2,23}$'),
 display_name text NOT NULL DEFAULT '', bio text NOT NULL DEFAULT '', public boolean NOT NULL DEFAULT true
);
CREATE TABLE handle_reservations (
 handle text PRIMARY KEY, account_id uuid REFERENCES accounts ON DELETE SET NULL,
 claimed_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE agents (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), account_id uuid NOT NULL UNIQUE REFERENCES accounts ON DELETE CASCADE
);
CREATE TABLE settings (
 account_id uuid PRIMARY KEY REFERENCES accounts ON DELETE CASCADE,
 theme text NOT NULL DEFAULT 'system' CHECK(theme IN ('system','dark','light')),
 notifications boolean NOT NULL DEFAULT false
);
CREATE TABLE rooms (slug text PRIMARY KEY, title text NOT NULL);
INSERT INTO rooms VALUES ('lobby','Community lobby');
CREATE TABLE memberships (
 room text REFERENCES rooms, account_id uuid REFERENCES accounts ON DELETE CASCADE,
 PRIMARY KEY(room,account_id)
);
CREATE TABLE messages (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, room text NOT NULL REFERENCES rooms,
 account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
 content text NOT NULL CHECK(char_length(content) BETWEEN 1 AND 4000), created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX messages_page ON messages(room,id DESC);
CREATE TABLE notes (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
 title text NOT NULL CHECK(char_length(title) <= 120), body text NOT NULL CHECK(char_length(body) <= 65536),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX notes_owner ON notes(account_id,updated_at DESC,id);
CREATE TABLE blocks (
 account_id uuid REFERENCES accounts ON DELETE CASCADE, blocked_id uuid REFERENCES accounts ON DELETE CASCADE,
 PRIMARY KEY(account_id,blocked_id), CHECK(account_id <> blocked_id)
);
CREATE TABLE reports (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(), account_id uuid REFERENCES accounts ON DELETE CASCADE,
 target_id uuid REFERENCES accounts ON DELETE SET NULL, message_id bigint REFERENCES messages ON DELETE SET NULL,
 reason text NOT NULL CHECK(char_length(reason) BETWEEN 1 AND 1000),
 status text NOT NULL DEFAULT 'open' CHECK(status IN ('open','resolved')),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE rate_limits (
 bucket text PRIMARY KEY, hits integer NOT NULL, expires_at timestamptz NOT NULL
);
CREATE TABLE idempotency (
 account_id uuid REFERENCES accounts ON DELETE CASCADE, key text NOT NULL,
 fingerprint text NOT NULL, response jsonb NOT NULL, status integer NOT NULL,
 expires_at timestamptz NOT NULL DEFAULT now() + interval '24 hours', PRIMARY KEY(account_id,key)
);
CREATE TABLE identity_audit (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
 account_id uuid REFERENCES accounts ON DELETE SET NULL, action text NOT NULL,
 occurred_at timestamptz NOT NULL DEFAULT now()
);
