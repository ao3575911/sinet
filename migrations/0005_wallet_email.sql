CREATE TABLE wallet_challenges (
 account_id uuid PRIMARY KEY REFERENCES accounts ON DELETE CASCADE,
 nonce text NOT NULL CHECK(nonce ~ '^[0-9a-f]{32}$'),
 address text NOT NULL CHECK(address ~ '^0x[0-9a-f]{40}$'),
 origin text NOT NULL CHECK(octet_length(origin) <= 512),
 message text NOT NULL CHECK(octet_length(message) <= 2048),
 expires_at timestamptz NOT NULL, attempts integer NOT NULL DEFAULT 0 CHECK(attempts BETWEEN 0 AND 5),
 consumed boolean NOT NULL DEFAULT false
);
CREATE TABLE wallet_links (
 account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
 chain_id integer NOT NULL CHECK(chain_id = 1),
 address text NOT NULL CHECK(address ~ '^0x[0-9a-f]{40}$'),
 proof_hash text NOT NULL, verified_at timestamptz NOT NULL DEFAULT now(),
 PRIMARY KEY(account_id,chain_id,address), UNIQUE(chain_id,address)
);
CREATE TABLE email_drafts (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 account_id uuid NOT NULL REFERENCES accounts ON DELETE CASCADE,
 recipient text NOT NULL DEFAULT '' CHECK(octet_length(recipient) <= 254),
 subject text NOT NULL CHECK(octet_length(subject) <= 200),
 body text NOT NULL CHECK(octet_length(body) <= 65536),
 updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX email_drafts_owner ON email_drafts(account_id,updated_at DESC,id);
