package api

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"html/template"
	"math/big"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

func (s *Server) otpHash(nonce, code string) string {
	m := hmac.New(sha256.New, s.OTPKey)
	_, _ = m.Write([]byte(nonce + ":" + code))
	return hex.EncodeToString(m.Sum(nil))
}
func normalizedEmail(raw string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(raw))
	a, err := mail.ParseAddress(email)
	if err != nil || a.Address != email || len(email) > 254 || strings.ContainsAny(email, "\r\n") || !strings.Contains(email, ".") {
		return "", bad("invalid_email")
	}
	return email, nil
}
func (c *call) requestOTP() (any, error) {
	var in struct {
		Email string `json:"email"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	email, err := normalizedEmail(in.Email)
	if err != nil {
		return nil, err
	}
	if err = c.limit("otp-ip:"+ip(c.r), 10, time.Hour); err != nil {
		return nil, err
	}
	if err = c.limit("otp-email:"+email, 3, 15*time.Minute); err != nil {
		return nil, err
	}
	if _, err = c.tx.ExecContext(c.r.Context(), `SELECT pg_advisory_xact_lock(hashtextextended($1,1))`, email); err != nil {
		return nil, err
	}
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		return nil, err
	}
	code := fmt.Sprintf("%08d", n.Int64())
	nonce := random(32)
	if err = c.exec(`INSERT INTO email_otps(email,nonce,code_hash,expires_at) VALUES($1,$2,$3,now()+interval '5 minutes') ON CONFLICT(email) DO UPDATE SET nonce=excluded.nonce,code_hash=excluded.code_hash,expires_at=excluded.expires_at,attempts=0,consumed=false`, email, nonce, c.s.otpHash(nonce, code)); err != nil {
		return nil, err
	}
	if c.s.Mail == nil || c.s.Mail.Send(c.r.Context(), email, code) != nil {
		if err = c.exec(`UPDATE email_otps SET consumed=true WHERE email=$1`, email); err != nil {
			return nil, err
		}
		return nil, fail(503, "email_delivery_unavailable")
	}
	return map[string]string{"status": "If delivery is available, a code has been sent. Check your email."}, nil
}
func (c *call) verifyOTP() (any, error) {
	var in struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	email, err := normalizedEmail(in.Email)
	if err != nil {
		return nil, err
	}
	if err = c.limit("verify-ip:"+ip(c.r), 30, 15*time.Minute); err != nil {
		return nil, err
	}
	if err = c.limit("verify-email:"+email, 15, 15*time.Minute); err != nil {
		return nil, err
	}
	var nonce, expected string
	var expires time.Time
	var attempts int
	var consumed bool
	err = c.tx.QueryRowContext(c.r.Context(), `SELECT nonce,code_hash,expires_at,attempts,consumed FROM email_otps WHERE email=$1 FOR UPDATE`, email).Scan(&nonce, &expected, &expires, &attempts, &consumed)
	if err == sql.ErrNoRows {
		return nil, fail(401, "invalid_or_expired_code")
	}
	if err != nil {
		return nil, err
	}
	if err = c.exec(`UPDATE email_otps SET attempts=attempts+1 WHERE email=$1`, email); err != nil {
		return nil, err
	}
	if consumed || attempts >= 5 || !expires.After(time.Now()) || len(in.Code) != 8 || !hmac.Equal([]byte(expected), []byte(c.s.otpHash(nonce, in.Code))) {
		return nil, fail(401, "invalid_or_expired_code")
	}
	if err = c.exec(`UPDATE email_otps SET consumed=true WHERE email=$1`, email); err != nil {
		return nil, err
	}
	var account string
	err = c.tx.QueryRowContext(c.r.Context(), `INSERT INTO accounts(email) VALUES($1) ON CONFLICT(email) DO UPDATE SET email=excluded.email RETURNING id`, email).Scan(&account)
	if err != nil {
		return nil, err
	}
	for _, q := range []string{`INSERT INTO profiles(account_id) VALUES($1) ON CONFLICT DO NOTHING`, `INSERT INTO agents(account_id) VALUES($1) ON CONFLICT DO NOTHING`, `INSERT INTO settings(account_id) VALUES($1) ON CONFLICT DO NOTHING`} {
		if err = c.exec(q, account); err != nil {
			return nil, err
		}
	}
	token := random(32)
	if err = c.exec(`INSERT INTO sessions(token_hash,account_id) VALUES($1,$2)`, hash(token), account); err != nil {
		return nil, err
	}
	if err = c.metric("verified_signin"); err != nil {
		return nil, err
	}
	return map[string]any{"token": token, "expires_in": 86400, "account_id": account}, nil
}
func reserved(handle string) bool {
	for _, name := range strings.Fields("admin administrator root system support help security official sinet api www mail email postmaster abuse moderator staff null undefined lobby agent connectors settings login logout me") {
		if strings.ReplaceAll(handle, "_", "") == strings.ReplaceAll(name, "_", "") {
			return true
		}
	}
	return false
}
func normalizeHandle(raw string) (string, error) {
	h := strings.ToLower(strings.TrimSpace(raw))
	if !handlePattern.MatchString(h) || reserved(h) {
		return "", bad("handle_must_be_3_to_24_ascii_chars_and_not_reserved")
	}
	return h, nil
}
func (c *call) availability() (any, error) {
	h, err := normalizeHandle(c.r.PathValue("name"))
	if err != nil {
		return nil, err
	}
	var taken bool
	err = c.tx.QueryRowContext(c.r.Context(), `SELECT EXISTS(SELECT 1 FROM handle_reservations WHERE handle=$1)`, h).Scan(&taken)
	return map[string]any{"handle": h, "available": !taken}, err
}
func (c *call) claim() (any, error) {
	var in struct {
		Handle string `json:"handle"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	h, err := normalizeHandle(in.Handle)
	if err != nil {
		return nil, err
	}
	if err = c.limit("claim:"+c.user, 10, time.Hour); err != nil {
		return nil, err
	}
	var current sql.NullString
	err = c.tx.QueryRowContext(c.r.Context(), `SELECT handle FROM profiles WHERE account_id=$1`, c.user).Scan(&current)
	if err != nil {
		return nil, err
	}
	if current.Valid {
		return nil, fail(409, "handle_already_claimed")
	}
	var claimed string
	err = c.tx.QueryRowContext(c.r.Context(), `INSERT INTO handle_reservations(handle,account_id) VALUES($1,$2) ON CONFLICT DO NOTHING RETURNING handle`, h, c.user).Scan(&claimed)
	if err == sql.ErrNoRows {
		return nil, fail(409, "handle_taken")
	}
	if err != nil {
		return nil, err
	}
	if err = c.exec(`UPDATE profiles SET handle=$1 WHERE account_id=$2`, h, c.user); err != nil {
		return nil, err
	}
	if err = c.exec(`INSERT INTO identity_audit(account_id,action) VALUES($1,'handle_claim')`, c.user); err != nil {
		return nil, err
	}
	if err = c.metric("handle_claimed"); err != nil {
		return nil, err
	}
	return map[string]string{"handle": h, "profile_url": c.s.PublicURL + "/u/" + h}, nil
}
func (c *call) me() (any, error) {
	return c.one(`SELECT json_build_object('id',a.id,'email',a.email,'handle',p.handle,'display_name',p.display_name,'bio',p.bio,'public',p.public,'agent_id',g.id,'model_available',$2::boolean,'email_alias',json_build_object('status','unprovisioned','address',null),'settings',json_build_object('theme',s.theme,'notifications',s.notifications),'profile_url',CASE WHEN p.handle IS NOT NULL THEN $3||'/u/'||p.handle END) FROM accounts a JOIN profiles p ON p.account_id=a.id JOIN agents g ON g.account_id=a.id JOIN settings s ON s.account_id=a.id WHERE a.id=$1`, c.user, c.s.Model != nil, c.s.PublicURL)
}
func (c *call) patchMe() (any, error) {
	var in struct {
		DisplayName   *string `json:"display_name"`
		Bio           *string `json:"bio"`
		Public        *bool   `json:"public"`
		Theme         *string `json:"theme"`
		Notifications *bool   `json:"notifications"`
	}
	if err := decode(c.r, &in); err != nil {
		return nil, err
	}
	if in.DisplayName != nil && utf8.RuneCountInString(*in.DisplayName) > 80 || in.Bio != nil && utf8.RuneCountInString(*in.Bio) > 280 {
		return nil, bad("profile_too_long")
	}
	if in.Theme != nil && *in.Theme != "system" && *in.Theme != "dark" && *in.Theme != "light" {
		return nil, bad("invalid_theme")
	}
	if err := c.exec(`UPDATE profiles SET display_name=COALESCE($2,display_name),bio=COALESCE($3,bio),public=COALESCE($4,public) WHERE account_id=$1`, c.user, in.DisplayName, in.Bio, in.Public); err != nil {
		return nil, err
	}
	if err := c.exec(`UPDATE settings SET theme=COALESCE($2,theme),notifications=COALESCE($3,notifications) WHERE account_id=$1`, c.user, in.Theme, in.Notifications); err != nil {
		return nil, err
	}
	return c.me()
}
func (c *call) publicProfile() (any, error) {
	return c.one(`SELECT json_build_object('handle',handle,'display_name',display_name,'bio',bio,'profile_url',$2||'/u/'||handle) FROM profiles WHERE handle=$1 AND public`, strings.ToLower(c.r.PathValue("handle")), c.s.PublicURL)
}
func (c *call) users() (any, error) {
	limit, _, err := page(c.r)
	if err != nil {
		return nil, err
	}
	off, err := offset(c.r)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(c.r.URL.Query().Get("q"))
	if len(q) > 80 {
		return nil, bad("query_too_long")
	}
	return c.many(`SELECT json_build_object('handle',p.handle,'display_name',p.display_name,'bio',p.bio) FROM profiles p WHERE p.public AND p.handle IS NOT NULL AND (strpos(p.handle,$1)>0 OR strpos(lower(p.display_name),$1)>0) AND NOT EXISTS(SELECT 1 FROM blocks b WHERE (b.account_id=$2 AND b.blocked_id=p.account_id) OR (b.blocked_id=$2 AND b.account_id=p.account_id)) ORDER BY p.handle LIMIT $3 OFFSET $4`, q, c.user, limit, off)
}

var profileTemplate = template.Must(template.New("profile").Parse(`<!doctype html><html lang="en"><meta charset="utf-8"><meta name="viewport" content="width=device-width"><title>@{{.Handle}} · SINet</title><meta property="og:title" content="@{{.Handle}} · SINet"><meta property="og:description" content="{{.Bio}}"><meta property="og:url" content="{{.URL}}"><meta name="description" content="{{.Bio}}"><body><main><h1>@{{.Handle}}</h1><h2>{{.Name}}</h2><p>{{.Bio}}</p><p>SINet · Your terminal. Your network.</p><p>Find this person: <code>SINet user {{.Handle}}</code></p></main></body></html>`))

func (s *Server) webProfile(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	var p struct{ Handle, Name, Bio, URL string }
	err := s.DB.QueryRowContext(ctx, `SELECT handle,display_name,bio FROM profiles WHERE handle=$1 AND public`, strings.ToLower(r.PathValue("handle"))).Scan(&p.Handle, &p.Name, &p.Bio)
	if err == sql.ErrNoRows {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		output(w, 503, map[string]string{"error": "database_unavailable"})
		return
	}
	p.URL = s.PublicURL + "/u/" + p.Handle
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = profileTemplate.Execute(w, p)
}
func (c *call) logout() (any, error) {
	err := c.exec(`DELETE FROM sessions WHERE token_hash=$1 AND account_id=$2`, hash(strings.TrimPrefix(c.r.Header.Get("Authorization"), "Bearer ")), c.user)
	return map[string]bool{"logged_out": true}, err
}
