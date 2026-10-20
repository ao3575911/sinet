package api

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func (c *call) emailStatus() (any, error) {
	return c.one(`SELECT json_build_object('sign_in_email',email,'sign_in_email_verified',true,'provider','gmail_web','inbox_url','https://mail.google.com/mail/u/0/#inbox','mailbox_sync_status','not_connected','drafts_available',true,'send_method','review_in_gmail','alias',json_build_object('status','unprovisioned','address',null)) FROM accounts WHERE id=$1`, c.user)
}

func (c *call) emailDrafts() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	return c.many(`SELECT json_build_object('id',id,'recipient',recipient,'subject',subject,'body',body,'updated_at',updated_at,'status','draft') FROM email_drafts WHERE account_id=$1 ORDER BY updated_at DESC,id LIMIT $2 OFFSET $3`, c.user, limit, off)
}

func (c *call) emailDraft() (any, error) {
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_draft_id")
	}
	return c.one(`SELECT json_build_object('id',id,'recipient',recipient,'subject',subject,'body',body,'updated_at',updated_at,'status','draft') FROM email_drafts WHERE account_id=$1 AND id=$2`, c.user, c.r.PathValue("id"))
}

func emailText(value string, multiline bool) bool {
	if !utf8.ValidString(value) {
		return false
	}
	for _, r := range value {
		if unicode.Is(unicode.Cf, r) || (unicode.IsControl(r) && !(multiline && (r == '\n' || r == '\t'))) {
			return false
		}
	}
	return true
}

func (c *call) saveEmailDraft() (any, error) {
	var in struct {
		Recipient string `json:"recipient"`
		Subject   string `json:"subject"`
		Body      string `json:"body"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	in.Body = strings.ReplaceAll(in.Body, "\r\n", "\n")
	if len(in.Subject) > 200 || len(in.Body) > 65536 || !emailText(in.Subject, false) || !emailText(in.Body, true) {
		return nil, bad("invalid_draft_text")
	}
	if in.Recipient != "" {
		// Preserve recipient local-part case; an address is not a SINet login identity.
		trimmed := strings.TrimSpace(in.Recipient)
		if len(trimmed) > 254 || !emailText(trimmed, false) {
			return nil, bad("invalid_recipient")
		}
		if _, err := normalizedEmail(trimmed); err != nil {
			return nil, bad("invalid_recipient")
		}
		in.Recipient = trimmed
	}
	if c.r.Method == "POST" {
		var count int
		if err := c.tx.QueryRowContext(c.r.Context(), `SELECT count(*) FROM email_drafts WHERE account_id=$1`, c.user).Scan(&count); err != nil {
			return nil, err
		}
		if count >= 100 {
			return nil, fail(429, "email_draft_limit_100")
		}
		return c.one(`INSERT INTO email_drafts(account_id,recipient,subject,body) VALUES($1,$2,$3,$4) RETURNING json_build_object('id',id,'recipient',recipient,'subject',subject,'body',body,'updated_at',updated_at,'status','draft')`, c.user, in.Recipient, in.Subject, in.Body)
	}
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_draft_id")
	}
	return c.one(`UPDATE email_drafts SET recipient=$3,subject=$4,body=$5,updated_at=now() WHERE account_id=$1 AND id=$2 RETURNING json_build_object('id',id,'recipient',recipient,'subject',subject,'body',body,'updated_at',updated_at,'status','draft')`, c.user, c.r.PathValue("id"), in.Recipient, in.Subject, in.Body)
}

func (c *call) deleteEmailDraft() (any, error) {
	if !uuidLike(c.r.PathValue("id")) {
		return nil, bad("invalid_draft_id")
	}
	var id string
	err := c.tx.QueryRowContext(c.r.Context(), `DELETE FROM email_drafts WHERE account_id=$1 AND id=$2 RETURNING id`, c.user, c.r.PathValue("id")).Scan(&id)
	return map[string]string{"deleted": id}, err
}
