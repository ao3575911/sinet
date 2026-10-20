package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
	"sinet/internal/mail"
	"sinet/internal/model"
)

type Server struct {
	DB        *sql.DB
	Mail      mail.Sender
	Model     model.Provider
	PublicURL string
	OTPKey    []byte
	mux       *http.ServeMux
}
type call struct {
	s    *Server
	tx   *sql.Tx
	r    *http.Request
	user string
}
type failure struct {
	Status int
	Code   string
}

func (e *failure) Error() string         { return e.Code }
func fail(status int, code string) error { return &failure{status, code} }
func bad(code string) error              { return fail(400, code) }
func random(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
func hash(v string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(v))) }
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(io.LimitReader(r.Body, 131073))
	d.DisallowUnknownFields()
	if err := d.Decode(v); err != nil {
		return bad("invalid_json")
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return bad("invalid_json")
	}
	return nil
}
func jsonBytes(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
func output(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(jsonBytes(v))
}
func New(db *sql.DB, sender mail.Sender, provider model.Provider, publicURL string) *Server {
	s := &Server{DB: db, Mail: sender, Model: provider, PublicURL: strings.TrimRight(publicURL, "/"), OTPKey: []byte(random(32)), mux: http.NewServeMux()}
	s.routes()
	return s
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'unsafe-inline'; frame-ancestors 'none'; base-uri 'none'")
	if r.ContentLength > 131072 {
		output(w, 413, map[string]string{"error": "request_too_large"})
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 131072)
	s.mux.ServeHTTP(w, r)
}
func (s *Server) route(pattern string, auth bool, fn func(*call) (any, error)) {
	s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 25*time.Second)
		defer cancel()
		r = r.WithContext(ctx)
		tx, err := s.DB.BeginTx(ctx, nil)
		if err != nil {
			output(w, 503, map[string]string{"error": "database_unavailable"})
			return
		}
		defer tx.Rollback()
		c := &call{s: s, tx: tx, r: r}
		mutation := r.Method != "GET" && r.Method != "HEAD"
		if auth {
			token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
			if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer ") || len(token) != 64 {
				output(w, 401, map[string]string{"error": "authentication_required"})
				return
			}
			err = tx.QueryRowContext(ctx, `SELECT account_id FROM sessions WHERE token_hash=$1 AND expires_at>now()`, hash(token)).Scan(&c.user)
			if err != nil {
				output(w, 401, map[string]string{"error": "authentication_required"})
				return
			}
			if mutation {
				if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, c.user); err != nil {
					output(w, 503, map[string]string{"error": "request_timeout"})
					return
				}
				var exists bool
				err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sessions WHERE token_hash=$1 AND expires_at>now())`, hash(token)).Scan(&exists)
				if err != nil || !exists {
					output(w, 401, map[string]string{"error": "authentication_required"})
					return
				}
			}
		}
		var key, fingerprint string
		if auth && mutation {
			key = r.Header.Get("Idempotency-Key")
			if len(key) < 16 || len(key) > 128 {
				output(w, 400, map[string]string{"error": "idempotency_key_required_16_to_128_chars"})
				return
			}
			raw, readErr := io.ReadAll(r.Body)
			if readErr != nil {
				output(w, 413, map[string]string{"error": "request_too_large"})
				return
			}
			r.Body = io.NopCloser(strings.NewReader(string(raw)))
			fingerprint = hash(r.Method + " " + r.URL.RequestURI() + " " + string(raw))
			var previous string
			var body []byte
			var status int
			err = tx.QueryRowContext(ctx, `SELECT fingerprint,response,status FROM idempotency WHERE account_id=$1 AND key=$2 AND expires_at>now()`, c.user, key).Scan(&previous, &body, &status)
			if err == nil {
				if previous != fingerprint {
					output(w, 409, map[string]string{"error": "idempotency_key_reused"})
				} else {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write(body)
				}
				return
			}
			if err != sql.ErrNoRows {
				output(w, 500, map[string]string{"error": "internal_error"})
				return
			}
			if err = c.limit("account:"+c.user, 120, time.Minute); err != nil {
				w.Header().Set("Retry-After", "60")
				_ = tx.Commit()
				output(w, 429, map[string]string{"error": "rate_limited"})
				return
			}
		}
		body, err := fn(c)
		status := 200
		if err != nil {
			var f *failure
			var pe *pq.Error
			if errors.As(err, &f) {
				status = f.Status
				body = map[string]string{"error": f.Code}
			} else if errors.As(err, &pe) && pe.Code == "23505" {
				_ = tx.Rollback()
				output(w, 409, map[string]string{"error": "conflict"})
				return
			} else if errors.Is(err, sql.ErrNoRows) {
				status = 404
				body = map[string]string{"error": "not_found"}
			} else {
				slog.Error("request failed", "method", r.Method, "route", pattern)
				output(w, 500, map[string]string{"error": "internal_error"})
				return
			}
		}
		if status == 429 {
			w.Header().Set("Retry-After", "60")
		}
		if auth && mutation && status < 400 && r.URL.Path != "/v1/me/delete" && r.URL.Path != "/v1/auth/logout" {
			_, err = tx.ExecContext(ctx, `INSERT INTO idempotency(account_id,key,fingerprint,response,status) VALUES($1,$2,$3,$4,$5) ON CONFLICT(account_id,key) DO UPDATE SET fingerprint=excluded.fingerprint,response=excluded.response,status=excluded.status,expires_at=excluded.expires_at`, c.user, key, fingerprint, string(jsonBytes(body)), status)
			if err != nil {
				output(w, 500, map[string]string{"error": "internal_error"})
				return
			}
		}
		if err = tx.Commit(); err != nil {
			output(w, 500, map[string]string{"error": "internal_error"})
			return
		}
		output(w, status, body)
	})
}
func (c *call) exec(q string, args ...any) error {
	_, err := c.tx.ExecContext(c.r.Context(), q, args...)
	return err
}
func (c *call) limit(bucket string, max int, window time.Duration) error {
	var hits int
	err := c.tx.QueryRowContext(c.r.Context(), `INSERT INTO rate_limits(bucket,hits,expires_at) VALUES($1,1,now()+$2*interval '1 second') ON CONFLICT(bucket) DO UPDATE SET hits=CASE WHEN rate_limits.expires_at<=now() THEN 1 ELSE rate_limits.hits+1 END, expires_at=CASE WHEN rate_limits.expires_at<=now() THEN excluded.expires_at ELSE rate_limits.expires_at END RETURNING hits`, hash(bucket), int(window.Seconds())).Scan(&hits)
	if err != nil {
		return err
	}
	if hits > max {
		return fail(429, "rate_limited")
	}
	return nil
}
func (c *call) metric(event string) error {
	return c.exec(`INSERT INTO metrics(event,count) VALUES($1,1) ON CONFLICT(day,event) DO UPDATE SET count=metrics.count+1`, event)
}
func (c *call) quota(kind string, max int) error {
	var count int
	err := c.tx.QueryRowContext(c.r.Context(), `INSERT INTO usage(account_id,kind,count) VALUES($1,$2,1) ON CONFLICT(account_id,day,kind) DO UPDATE SET count=usage.count+1 RETURNING count`, c.user, kind).Scan(&count)
	if err != nil {
		return err
	}
	if count > max {
		return fail(429, "daily_quota_exhausted")
	}
	return nil
}
func (c *call) activated() error {
	var handle sql.NullString
	err := c.tx.QueryRowContext(c.r.Context(), `SELECT handle FROM profiles WHERE account_id=$1`, c.user).Scan(&handle)
	if err != nil {
		return err
	}
	if !handle.Valid {
		return fail(403, "claim_handle_first")
	}
	return nil
}
func page(r *http.Request) (int, int64, error) {
	limit := 20
	var before int64
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 50 {
			return 0, 0, bad("invalid_limit")
		}
	}
	if raw := r.URL.Query().Get("before"); raw != "" {
		before, err = strconv.ParseInt(raw, 10, 64)
		if err != nil || before < 1 {
			return 0, 0, bad("invalid_cursor")
		}
	}
	return limit, before, nil
}
func offset(r *http.Request) (int, error) {
	n := 0
	var err error
	if v := r.URL.Query().Get("offset"); v != "" {
		n, err = strconv.Atoi(v)
	}
	if err != nil || n < 0 || n > 100000 {
		return 0, bad("invalid_offset")
	}
	return n, nil
}
func ip(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

var handlePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,23}$`)
