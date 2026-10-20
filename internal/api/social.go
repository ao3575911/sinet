package api

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

func (c *call) rooms() (any, error) {
	return c.many(`SELECT json_build_object('slug',slug,'title',title) FROM rooms ORDER BY slug LIMIT 50`)
}
func (c *call) join() (any, error) {
	if err := c.activated(); err != nil {
		return nil, err
	}
	var room string
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT slug FROM rooms WHERE slug=$1`, c.r.PathValue("slug")).Scan(&room)
	if err != nil {
		return nil, err
	}
	err = c.exec(`INSERT INTO memberships(room,account_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, room, c.user)
	return map[string]string{"joined": room}, err
}
func (c *call) member() error {
	var exists bool
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT EXISTS(SELECT 1 FROM memberships WHERE room=$1 AND account_id=$2)`, c.r.PathValue("slug"), c.user).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return fail(403, "join_room_first")
	}
	return nil
}
func (c *call) messages() (any, error) {
	if err := c.member(); err != nil {
		return nil, err
	}
	limit, before, err := page(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',m.id,'handle',p.handle,'content',m.content,'created_at',m.created_at) FROM messages m JOIN profiles p ON p.account_id=m.account_id WHERE m.room=$1 AND ($3::bigint=0 OR m.id<$3) AND NOT EXISTS(SELECT 1 FROM blocks b WHERE (b.account_id=$2 AND b.blocked_id=m.account_id) OR (b.blocked_id=$2 AND b.account_id=m.account_id)) ORDER BY m.id DESC LIMIT $4`, c.r.PathValue("slug"), c.user, before, limit)
}
func (c *call) postMessage() (any, error) {
	if err := c.activated(); err != nil {
		return nil, err
	}
	if err := c.member(); err != nil {
		return nil, err
	}
	var in struct {
		Content string `json:"content"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Content) == "" || utf8.RuneCountInString(in.Content) > 4000 {
		return nil, bad("message_must_be_1_to_4000_chars")
	}
	if err := c.quota("room_messages", 200); err != nil {
		return nil, err
	}
	return c.one(`INSERT INTO messages(room,account_id,content) VALUES($1,$2,$3) RETURNING json_build_object('id',id,'content',content,'created_at',created_at)`, c.r.PathValue("slug"), c.user, in.Content)
}
func (c *call) block() (any, error) {
	var in struct {
		Handle string `json:"handle"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	var target string
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT account_id FROM profiles WHERE handle=$1`, strings.ToLower(in.Handle)).Scan(&target)
	if err != nil {
		return nil, err
	}
	if target == c.user {
		return nil, bad("cannot_block_self")
	}
	err = c.exec(`INSERT INTO blocks(account_id,blocked_id) VALUES($1,$2) ON CONFLICT DO NOTHING`, c.user, target)
	return map[string]bool{"blocked": true}, err
}
func (c *call) unblock() (any, error) {
	err := c.exec(`DELETE FROM blocks WHERE account_id=$1 AND blocked_id=(SELECT account_id FROM profiles WHERE handle=$2)`, c.user, strings.ToLower(c.r.PathValue("handle")))
	return map[string]bool{"unblocked": true}, err
}
func (c *call) blocks() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('handle',p.handle) FROM blocks b JOIN profiles p ON p.account_id=b.blocked_id WHERE b.account_id=$1 ORDER BY p.handle LIMIT $2 OFFSET $3`, c.user, limit, off)
}
func (c *call) report() (any, error) {
	var in struct {
		Handle    string `json:"handle"`
		MessageID int64  `json:"message_id"`
		Reason    string `json:"reason"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Reason) == "" || utf8.RuneCountInString(in.Reason) > 1000 || (in.Handle == "" && in.MessageID == 0) {
		return nil, bad("invalid_report")
	}
	if err := c.quota("reports", 20); err != nil {
		return nil, err
	}
	var target any
	var message any
	if in.Handle != "" {
		var id string
		if err := c.tx.QueryRowContext(c.r.Context(), `SELECT account_id FROM profiles WHERE handle=$1`, strings.ToLower(in.Handle)).Scan(&id); err != nil {
			return nil, err
		}
		target = id
	}
	if in.MessageID != 0 {
		var id int64
		if err := c.tx.QueryRowContext(c.r.Context(), `SELECT m.id FROM messages m JOIN memberships j ON j.room=m.room AND j.account_id=$1 WHERE m.id=$2`, c.user, in.MessageID).Scan(&id); err != nil {
			return nil, err
		}
		message = id
	}
	return c.one(`INSERT INTO reports(account_id,target_id,message_id,reason) VALUES($1,$2,$3,$4) RETURNING json_build_object('id',id,'status',status)`, c.user, target, message, in.Reason)
}
func (c *call) notes() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',id,'title',title,'body',body,'updated_at',updated_at) FROM notes WHERE account_id=$1 ORDER BY updated_at DESC,id LIMIT $2 OFFSET $3`, c.user, limit, off)
}
func (c *call) note() (any, error) {
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_note_id")
	}
	return c.one(`SELECT json_build_object('id',id,'title',title,'body',body,'updated_at',updated_at) FROM notes WHERE account_id=$1 AND id=$2`, c.user, c.r.PathValue("id"))
}
func uuidLike(v string) bool {
	if len(v) != 36 {
		return false
	}
	for i, c := range v {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !strings.ContainsRune("0123456789abcdefABCDEF", c) {
			return false
		}
	}
	return true
}
func (c *call) saveNote() (any, error) {
	var in struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if utf8.RuneCountInString(in.Title) > 120 || len(in.Body) > 65536 {
		return nil, bad("note_too_long")
	}
	if c.r.Method == "POST" {
		var count int
		if err := c.tx.QueryRowContext(c.r.Context(), `SELECT count(*) FROM notes WHERE account_id=$1`, c.user).Scan(&count); err != nil {
			return nil, err
		}
		if count >= 100 {
			return nil, fail(429, "note_limit_100")
		}
		return c.one(`INSERT INTO notes(account_id,title,body) VALUES($1,$2,$3) RETURNING json_build_object('id',id,'title',title,'body',body,'updated_at',updated_at)`, c.user, in.Title, in.Body)
	}
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_note_id")
	}
	return c.one(`UPDATE notes SET title=$3,body=$4,updated_at=now() WHERE account_id=$1 AND id=$2 RETURNING json_build_object('id',id,'title',title,'body',body,'updated_at',updated_at)`, c.user, c.r.PathValue("id"), in.Title, in.Body)
}
func (c *call) deleteNote() (any, error) {
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_note_id")
	}
	var id string
	err := c.tx.QueryRowContext(c.r.Context(), `DELETE FROM notes WHERE account_id=$1 AND id=$2 RETURNING id`, c.user, c.r.PathValue("id")).Scan(&id)
	return map[string]string{"deleted": id}, err
}
func (c *call) export() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	me, err := c.me()
	if err != nil {
		return nil, err
	}
	result["account"] = me
	queries := map[string]string{
		"notes":          `SELECT row_to_json(x) FROM (SELECT id,title,body,updated_at FROM notes WHERE account_id=$1 ORDER BY id) x ORDER BY x.id`,
		"messages":       `SELECT row_to_json(x) FROM (SELECT id,room,content,created_at FROM messages WHERE account_id=$1 ORDER BY id) x ORDER BY x.id`,
		"blocks":         `SELECT json_build_object('handle',p.handle) FROM blocks b JOIN profiles p ON p.account_id=b.blocked_id WHERE b.account_id=$1 ORDER BY p.handle`,
		"reports":        `SELECT row_to_json(x) FROM (SELECT id,target_id,message_id,reason,status FROM reports WHERE account_id=$1) x ORDER BY x.id`,
		"installs":       `SELECT json_build_object('connector_id',connector_id,'version',version) FROM installs WHERE account_id=$1 ORDER BY connector_id`,
		"grants":         `SELECT json_build_object('connector_id',connector_id,'operation',operation) FROM grants WHERE account_id=$1 ORDER BY connector_id,operation`,
		"agent_messages": `SELECT json_build_object('id',id,'role',role,'content',content,'created_at',created_at) FROM agent_messages WHERE account_id=$1 ORDER BY id`,
		"tool_audit":     `SELECT json_build_object('id',id,'connector_id',connector_id,'version',version,'operation',operation,'allowed',allowed,'outcome',outcome) FROM tool_audit WHERE account_id=$1 ORDER BY id`,
	}
	hasMore := false
	for name, q := range queries {
		value, e := c.many(q+" LIMIT $2 OFFSET $3", c.user, limit, off)
		if e != nil {
			return nil, e
		}
		result[name] = value
		if len(value.(map[string]any)["items"].([]json.RawMessage)) == limit {
			hasMore = true
		}
	}
	var next any
	if hasMore {
		next = off + limit
	}
	result["pagination"] = map[string]any{"limit": limit, "offset": off, "next_offset": next}
	return result, nil
}
func (c *call) deleteAccount() (any, error) {
	var in struct {
		Confirm string `json:"confirm"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if in.Confirm != "delete my account" {
		return nil, bad("explicit_deletion_confirmation_required")
	}
	if err := c.exec(`DELETE FROM email_otps WHERE email=(SELECT email FROM accounts WHERE id=$1)`, c.user); err != nil {
		return nil, err
	}
	if err := c.exec(`UPDATE connectors SET status='rejected',manifest=jsonb_set(manifest,'{publisher}','"deleted publisher"') WHERE publisher_id=$1`, c.user); err != nil {
		return nil, err
	}
	if err := c.exec(`INSERT INTO identity_audit(account_id,action) VALUES($1,'account_delete')`, c.user); err != nil {
		return nil, err
	}
	if err := c.exec(`DELETE FROM accounts WHERE id=$1`, c.user); err != nil {
		return nil, err
	}
	return map[string]bool{"deleted": true}, nil
}
