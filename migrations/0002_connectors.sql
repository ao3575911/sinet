CREATE TABLE connectors (
 id text NOT NULL, version text NOT NULL, publisher_id uuid REFERENCES accounts ON DELETE SET NULL,
 manifest jsonb NOT NULL, status text NOT NULL CHECK(status IN ('review','approved','rejected')),
 review_reference text, submitted_at timestamptz NOT NULL DEFAULT now(), PRIMARY KEY(id,version)
);
CREATE TABLE installs (
 account_id uuid REFERENCES accounts ON DELETE CASCADE, connector_id text NOT NULL, version text NOT NULL,
 PRIMARY KEY(account_id,connector_id), FOREIGN KEY(connector_id,version) REFERENCES connectors(id,version)
);
CREATE TABLE grants (
 account_id uuid, connector_id text, operation text NOT NULL,
 PRIMARY KEY(account_id,connector_id,operation),
 FOREIGN KEY(account_id,connector_id) REFERENCES installs(account_id,connector_id) ON DELETE CASCADE
);
CREATE TABLE consents (
 token_hash text PRIMARY KEY, account_id uuid REFERENCES accounts ON DELETE CASCADE,
 fingerprint text NOT NULL, expires_at timestamptz NOT NULL DEFAULT now()+interval '5 minutes', consumed boolean NOT NULL DEFAULT false
);
CREATE TABLE tool_audit (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, account_id uuid REFERENCES accounts ON DELETE CASCADE,
 connector_id text NOT NULL, version text NOT NULL, operation text NOT NULL,
 allowed boolean NOT NULL, outcome text NOT NULL, occurred_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX audit_owner ON tool_audit(account_id,id DESC);
CREATE TABLE agent_messages (
 id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, account_id uuid REFERENCES accounts ON DELETE CASCADE,
 role text NOT NULL CHECK(role IN ('user','assistant')), content text NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE usage (
 account_id uuid REFERENCES accounts ON DELETE CASCADE, day date NOT NULL DEFAULT CURRENT_DATE,
 kind text NOT NULL, count integer NOT NULL DEFAULT 0, PRIMARY KEY(account_id,day,kind)
);
CREATE TABLE metrics (
 day date NOT NULL DEFAULT CURRENT_DATE, event text NOT NULL, count bigint NOT NULL DEFAULT 0,
 PRIMARY KEY(day,event)
);
