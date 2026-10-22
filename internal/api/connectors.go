package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"sinet/internal/connectors"
	"sinet/internal/githubpublic"
	"sinet/internal/model"
	"sinet/internal/strictjson"
	"sinet/prompts"
)

func (c *call) connectors() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('manifest',manifest,'status',status,'review_reference',review_reference) FROM connectors WHERE status='approved' ORDER BY id,submitted_at DESC,version LIMIT $1 OFFSET $2`, limit, off)
}
func (c *call) connector() (any, error) {
	return c.one(`SELECT json_build_object('manifest',manifest,'status',status,'review_reference',review_reference) FROM connectors WHERE id=$1 AND status='approved' AND ($2='' OR version=$2) ORDER BY submitted_at DESC,version DESC LIMIT 1`, c.r.PathValue("id"), c.r.URL.Query().Get("version"))
}
func (c *call) submissions() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('manifest',manifest,'status',status,'review_reference',review_reference) FROM connectors WHERE publisher_id=$1 ORDER BY submitted_at DESC,id,version LIMIT $2 OFFSET $3`, c.user, limit, off)
}
func (c *call) submitConnector() (any, error) {
	if err := c.activated(); err != nil {
		return nil, err
	}
	var raw json.RawMessage
	if err := decode(c.r, &raw); err != nil {
		return nil, err
	}
	m, err := connectors.Parse(raw)
	if err != nil {
		return nil, bad("invalid_manifest")
	}
	var handle string
	if err = c.tx.QueryRowContext(c.r.Context(), `SELECT handle FROM profiles WHERE account_id=$1`, c.user).Scan(&handle); err != nil {
		return nil, err
	}
	if m.Publisher != handle {
		return nil, fail(403, "publisher_must_match_your_handle")
	}
	// Account locks do not serialize competing publishers of the same ID.
	if _, err = c.tx.ExecContext(c.r.Context(), `SELECT pg_advisory_xact_lock(hashtextextended($1,2))`, m.ID); err != nil {
		return nil, err
	}
	var occupied bool
	if err = c.tx.QueryRowContext(c.r.Context(), `SELECT EXISTS(SELECT 1 FROM connectors WHERE id=$1 AND (publisher_id IS NULL OR publisher_id<>$2))`, m.ID, c.user).Scan(&occupied); err != nil {
		return nil, err
	}
	if occupied {
		return nil, fail(409, "connector_id_owned_by_another_publisher")
	}
	if err = c.quota("submissions", 10); err != nil {
		return nil, err
	}
	var id string
	err = c.tx.QueryRowContext(c.r.Context(), `INSERT INTO connectors(id,version,publisher_id,manifest,status) VALUES($1,$2,$3,$4,'review') ON CONFLICT DO NOTHING RETURNING id`, m.ID, m.Version, c.user, string(raw)).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, fail(409, "immutable_version_already_exists")
	}
	if err != nil {
		return nil, err
	}
	return map[string]string{"id": id, "version": m.Version, "status": "review"}, nil
}
func (c *call) installs() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',i.connector_id,'version',i.version,'status',c.status,'grants',COALESCE((SELECT json_agg(g.operation ORDER BY g.operation) FROM grants g WHERE g.account_id=i.account_id AND g.connector_id=i.connector_id),'[]'::json)) FROM installs i JOIN connectors c ON c.id=i.connector_id AND c.version=i.version WHERE i.account_id=$1 ORDER BY i.connector_id LIMIT $2 OFFSET $3`, c.user, limit, off)
}
func (c *call) install() (any, error) {
	if err := c.activated(); err != nil {
		return nil, err
	}
	var in struct {
		Version string `json:"version"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	var approved bool
	if err := c.tx.QueryRowContext(c.r.Context(), `SELECT EXISTS(SELECT 1 FROM connectors WHERE id=$1 AND version=$2 AND status='approved')`, c.r.PathValue("id"), in.Version).Scan(&approved); err != nil {
		return nil, err
	}
	if !approved {
		return nil, fail(403, "reviewed_version_required")
	}
	var count int
	if err := c.tx.QueryRowContext(c.r.Context(), `SELECT count(*) FROM installs WHERE account_id=$1`, c.user).Scan(&count); err != nil {
		return nil, err
	}
	if count >= 50 {
		return nil, fail(429, "installation_limit_50")
	}
	var id string
	err := c.tx.QueryRowContext(c.r.Context(), `INSERT INTO installs(account_id,connector_id,version) VALUES($1,$2,$3) ON CONFLICT DO NOTHING RETURNING connector_id`, c.user, c.r.PathValue("id"), in.Version).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, fail(409, "already_installed_remove_before_version_change")
	}
	if err != nil {
		return nil, err
	}
	if err = c.metric("connector_installed"); err != nil {
		return nil, err
	}
	return map[string]string{"installed": id, "version": in.Version}, nil
}
func (c *call) uninstall() (any, error) {
	if err := c.exec(`DELETE FROM consents WHERE account_id=$1`, c.user); err != nil {
		return nil, err
	}
	err := c.exec(`DELETE FROM installs WHERE account_id=$1 AND connector_id=$2`, c.user, c.r.PathValue("id"))
	return map[string]bool{"uninstalled": true}, err
}
func (c *call) installed(id, version string) (connectors.Manifest, error) {
	var raw []byte
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT c.manifest FROM installs i JOIN connectors c ON c.id=i.connector_id AND c.version=i.version WHERE i.account_id=$1 AND i.connector_id=$2 AND i.version=$3 AND c.status='approved'`, c.user, id, version).Scan(&raw)
	if err == sql.ErrNoRows {
		return connectors.Manifest{}, fail(403, "approved_pinned_installation_required")
	}
	if err != nil {
		return connectors.Manifest{}, err
	}
	m, err := connectors.Parse(raw)
	if err != nil {
		return m, fail(403, "manifest_disabled")
	}
	return m, nil
}
func (c *call) grant() (any, error) {
	var in struct {
		Version   string `json:"version"`
		Operation string `json:"operation"`
		Approve   bool   `json:"approve"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if !in.Approve {
		return nil, bad("explicit_grant_approval_required")
	}
	m, err := c.installed(c.r.PathValue("id"), in.Version)
	if err != nil {
		return nil, err
	}
	if _, err = m.Operation(in.Operation); err != nil {
		return nil, bad("unknown_operation")
	}
	err = c.exec(`INSERT INTO grants(account_id,connector_id,operation) VALUES($1,$2,$3) ON CONFLICT DO NOTHING`, c.user, m.ID, in.Operation)
	return map[string]string{"granted": in.Operation, "connector": m.ID, "version": m.Version}, err
}
func (c *call) revoke() (any, error) {
	if err := c.exec(`DELETE FROM consents WHERE account_id=$1`, c.user); err != nil {
		return nil, err
	}
	err := c.exec(`DELETE FROM grants WHERE account_id=$1 AND connector_id=$2 AND operation=$3`, c.user, c.r.PathValue("id"), c.r.PathValue("operation"))
	return map[string]bool{"revoked": true}, err
}

type toolRequest struct {
	Connector string          `json:"connector"`
	Version   string          `json:"version"`
	Operation string          `json:"operation"`
	Arguments json.RawMessage `json:"arguments"`
	Consent   string          `json:"consent,omitempty"`
}

func (t toolRequest) fingerprint() string {
	var v any
	_ = strictjson.Decode(t.Arguments, &v)
	return hash(t.Connector + ":" + t.Version + ":" + t.Operation + ":" + string(jsonBytes(v)))
}
func (c *call) authorizeTool(t toolRequest) (connectors.Manifest, connectors.Operation, error) {
	m, err := c.installed(t.Connector, t.Version)
	if err != nil {
		return m, connectors.Operation{}, err
	}
	op, err := m.Operation(t.Operation)
	if err != nil {
		return m, op, bad("unknown_operation")
	}
	var granted bool
	err = c.tx.QueryRowContext(c.r.Context(), `SELECT EXISTS(SELECT 1 FROM grants WHERE account_id=$1 AND connector_id=$2 AND operation=$3)`, c.user, t.Connector, t.Operation).Scan(&granted)
	if err != nil {
		return m, op, err
	}
	if !granted {
		return m, op, fail(403, "operation_grant_required")
	}
	if m.ID == "github-public" && !githubpublic.ValidInput(t.Arguments) {
		return m, op, bad("invalid_tool_arguments")
	}
	if err = op.Validate(t.Arguments); err != nil {
		return m, op, bad("invalid_tool_arguments")
	}
	return m, op, nil
}
func needsConsent(m connectors.Manifest, op connectors.Operation) bool {
	return m.Runtime == "remote_https" || op.Risk != "read"
}
func (c *call) consent() (any, error) {
	var in struct {
		Tool    toolRequest `json:"tool"`
		Approve bool        `json:"approve"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if !in.Approve {
		return nil, bad("explicit_call_approval_required")
	}
	m, op, err := c.authorizeTool(in.Tool)
	if err != nil {
		return nil, err
	}
	if !needsConsent(m, op) {
		return nil, bad("this_read_operation_needs_no_call_consent")
	}
	token := random(32)
	if err = c.exec(`INSERT INTO consents(token_hash,account_id,fingerprint) VALUES($1,$2,$3)`, hash(token), c.user, in.Tool.fingerprint()); err != nil {
		return nil, err
	}
	disclosure := connectors.Disclosure(m, op, in.Tool.Arguments)
	// Preserve the existing consent destination host field; recipients carry URLs.
	destination := "builtin"
	if u, e := url.Parse(m.EndpointURL); e == nil && u.Host != "" {
		destination = u.Host
	}
	disclosure["destination"] = destination
	disclosure["consent"] = token
	disclosure["expires_in"] = 300
	return disclosure, nil
}
func (c *call) audit(t toolRequest, allowed bool, outcome string) error {
	return c.exec(`INSERT INTO tool_audit(account_id,connector_id,version,operation,allowed,outcome) VALUES($1,$2,$3,$4,$5,$6)`, c.user, t.Connector, t.Version, t.Operation, allowed, outcome)
}
func (c *call) runTool(t toolRequest) (any, error) {
	if len(t.Connector) > 63 || len(t.Version) > 32 || len(t.Operation) > 64 {
		return nil, bad("invalid_tool_request")
	}
	if err := c.quota("tools", 100); err != nil {
		return nil, err
	}
	m, op, err := c.authorizeTool(t)
	if err != nil {
		if e := c.audit(t, false, "authorization_denied"); e != nil {
			return nil, e
		}
		return nil, err
	}
	if needsConsent(m, op) {
		var token string
		err = c.tx.QueryRowContext(c.r.Context(), `UPDATE consents SET consumed=true WHERE token_hash=$1 AND account_id=$2 AND fingerprint=$3 AND NOT consumed AND expires_at>clock_timestamp() RETURNING token_hash`, hash(t.Consent), c.user, t.fingerprint()).Scan(&token)
		if err == sql.ErrNoRows {
			if e := c.audit(t, false, "consent_required"); e != nil {
				return nil, e
			}
			return nil, fail(403, "fresh_call_consent_required")
		}
		if err != nil {
			return nil, err
		}
	}
	result, err := c.s.executeTool(c.r.Context(), m, op, t.Arguments)
	if err != nil {
		if e := c.audit(t, true, "execution_failed"); e != nil {
			return nil, e
		}
		var remote *connectors.RemoteFailure
		if errors.As(err, &remote) {
			return nil, fail(remote.Status, remote.Code)
		}
		return nil, fail(502, "connector_unavailable_or_destination_denied")
	}
	if err = c.audit(t, true, "success"); err != nil {
		return nil, err
	}
	if err = c.metric("connector_used"); err != nil {
		return nil, err
	}
	return map[string]any{"mode": "deterministic_tool", "connector": m.ID, "version": m.Version, "operation": op.Name, "output": result}, nil
}
func (c *call) agentMessage() (any, error) {
	var in struct {
		Message    string       `json:"message"`
		AllowModel bool         `json:"allow_model"`
		Tool       *toolRequest `json:"tool"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if err := c.activated(); err != nil {
		return nil, err
	}
	if in.Tool != nil {
		if in.Message != "" || in.AllowModel {
			return nil, bad("choose_chat_or_explicit_tool")
		}
		return c.runTool(*in.Tool)
	}
	if strings.TrimSpace(in.Message) == "" || len(in.Message) > 4096 {
		return nil, bad("message_must_be_1_to_4096_bytes")
	}
	if c.s.Model == nil {
		return map[string]string{"mode": "no_model", "status": "AI unavailable: no model provider configured. Explicit deterministic connector tools remain available."}, nil
	}
	if !in.AllowModel {
		return nil, fail(403, "explicit_model_disclosure_approval_required")
	}
	if err := c.quota("model", 40); err != nil {
		return nil, err
	}
	messages := []model.Message{{Role: "system", Content: prompts.AgentPolicy + "\nYou are a chat-only model. Tools execute only through explicit authenticated API requests; never claim you executed a tool."}}
	rows, err := c.tx.QueryContext(c.r.Context(), `SELECT role,content FROM (SELECT id,role,content FROM agent_messages WHERE account_id=$1 ORDER BY id DESC LIMIT 20) x ORDER BY id`, c.user)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var msg model.Message
		if err = rows.Scan(&msg.Role, &msg.Content); err != nil {
			rows.Close()
			return nil, err
		}
		messages = append(messages, msg)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	messages = append(messages, model.Message{Role: "user", Content: in.Message})
	answer, err := c.s.Model.Chat(c.r.Context(), messages)
	if err != nil {
		return nil, fail(502, "model_unavailable")
	}
	if len(answer) > 16384 {
		return nil, fail(502, "model_output_limit")
	}
	if err = c.exec(`INSERT INTO agent_messages(account_id,role,content) VALUES($1,'user',$2),($1,'assistant',$3)`, c.user, in.Message, answer); err != nil {
		return nil, err
	}
	if err = c.exec(`DELETE FROM agent_messages WHERE account_id=$1 AND id NOT IN (SELECT id FROM agent_messages WHERE account_id=$1 ORDER BY id DESC LIMIT 200)`, c.user); err != nil {
		return nil, err
	}
	if err = c.metric("model_chat"); err != nil {
		return nil, err
	}
	return map[string]any{"mode": "model", "message": answer}, nil
}
func (c *call) agentHistory() (any, error) {
	limit, before, err := page(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',id,'role',role,'content',content,'created_at',created_at) FROM agent_messages WHERE account_id=$1 AND ($2::bigint=0 OR id<$2) ORDER BY id DESC LIMIT $3`, c.user, before, limit)
}
func (c *call) toolAudit() (any, error) {
	limit, before, err := page(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',id,'connector',connector_id,'version',version,'operation',operation,'allowed',allowed,'outcome',outcome,'occurred_at',occurred_at) FROM tool_audit WHERE account_id=$1 AND ($2::bigint=0 OR id<$2) ORDER BY id DESC LIMIT $3`, c.user, before, limit)
}
func (c *call) usage() (any, error) {
	return c.many(`SELECT json_build_object('day',day,'kind',kind,'count',count) FROM usage WHERE account_id=$1 ORDER BY day DESC,kind LIMIT 50`, c.user)
}
