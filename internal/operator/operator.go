package operator

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"sinet/internal/connectors"
)

func Run(ctx context.Context, db *sql.DB, args []string, out io.Writer) error {
	if len(args) == 0 {
		return errors.New("operator command required: review, reports, metrics, maintenance")
	}
	switch args[0] {
	case "review":
		if len(args) == 2 && args[1] == "list" {
			return printRows(ctx, db, out, `SELECT json_build_object('id',id,'version',version,'status',status,'manifest',manifest) FROM connectors WHERE status='review' ORDER BY submitted_at LIMIT 50`)
		}
		if len(args) != 6 || (args[1] != "approve" && args[1] != "reject") || args[5] != "--confirm" {
			return errors.New("usage: server review approve|reject ID VERSION OWNERSHIP_AND_REVIEW_REFERENCE --confirm")
		}
		if len(args[4]) < 12 || len(args[4]) > 500 {
			return errors.New("a substantive ownership and trust review reference is required")
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var raw []byte
		var status string
		if err = tx.QueryRowContext(ctx, `SELECT manifest,status FROM connectors WHERE id=$1 AND version=$2 FOR UPDATE`, args[2], args[3]).Scan(&raw, &status); err != nil {
			return err
		}
		if args[1] == "approve" {
			if status != "review" {
				return errors.New("only review submissions may be approved")
			}
			m, err := connectors.Parse(raw)
			if err != nil {
				return err
			}
			if m.Runtime != "remote_https" {
				return errors.New("only bundled audited builtins execute; submitted builtins cannot be approved")
			}
			if !strings.Contains(args[4], "ownership:") || !strings.Contains(args[4], "review:") {
				return errors.New("reference must identify ownership: evidence and review: evidence; verify them out of band first")
			}
		}
		next := "rejected"
		if args[1] == "approve" {
			next = "approved"
		}
		if _, err = tx.ExecContext(ctx, `UPDATE connectors SET status=$3,review_reference=$4 WHERE id=$1 AND version=$2`, args[2], args[3], next, args[4]); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		_, err = fmt.Fprintf(out, "%s %s %s\n", args[2], args[3], next)
		return err
	case "reports":
		if len(args) == 2 && args[1] == "list" {
			return printRows(ctx, db, out, `SELECT json_build_object('id',id,'message_id',message_id,'target_id',target_id,'reason',reason,'created_at',created_at) FROM reports WHERE status='open' ORDER BY created_at LIMIT 50`)
		}
		if len(args) != 5 || args[1] != "resolve" || (args[3] != "keep" && args[3] != "remove-message") || args[4] != "--confirm" {
			return errors.New("usage: server reports resolve REPORT_UUID keep|remove-message --confirm")
		}
		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var message sql.NullInt64
		if err = tx.QueryRowContext(ctx, `SELECT message_id FROM reports WHERE id=$1 AND status='open' FOR UPDATE`, args[2]).Scan(&message); err != nil {
			return err
		}
		if args[3] == "remove-message" && message.Valid {
			if _, err = tx.ExecContext(ctx, `DELETE FROM messages WHERE id=$1`, message.Int64); err != nil {
				return err
			}
		}
		if _, err = tx.ExecContext(ctx, `UPDATE reports SET status='resolved' WHERE id=$1`, args[2]); err != nil {
			return err
		}
		if err = tx.Commit(); err != nil {
			return err
		}
		_, err = fmt.Fprintln(out, "report resolved")
		return err
	case "metrics":
		return printRows(ctx, db, out, `SELECT json_build_object('day',day,'event',event,'count',count) FROM metrics ORDER BY day DESC,event LIMIT 500`)
	case "maintenance":
		if len(args) != 2 || args[1] != "--confirm" {
			return errors.New("usage: server maintenance --confirm")
		}
		_, err := db.ExecContext(ctx, `DELETE FROM email_otps WHERE expires_at<now()-interval '1 day'; DELETE FROM sessions WHERE expires_at<now(); DELETE FROM consents WHERE expires_at<now(); DELETE FROM idempotency WHERE expires_at<now(); DELETE FROM rate_limits WHERE expires_at<now(); DELETE FROM agent_messages WHERE created_at<now()-interval '30 days'; DELETE FROM tool_audit WHERE occurred_at<now()-interval '90 days'; DELETE FROM messages WHERE created_at<now()-interval '90 days'; DELETE FROM usage WHERE day<CURRENT_DATE-90; DELETE FROM reports WHERE created_at<now()-interval '365 days'; DELETE FROM identity_audit WHERE occurred_at<now()-interval '365 days'`)
		if err == nil {
			_, err = fmt.Fprintln(out, "retention maintenance complete")
		}
		return err
	}
	return errors.New("unknown operator command")
}
func printRows(ctx context.Context, db *sql.DB, out io.Writer, q string) error {
	rows, err := db.QueryContext(ctx, q)
	if err != nil {
		return err
	}
	defer rows.Close()
	enc := json.NewEncoder(out)
	for rows.Next() {
		var raw json.RawMessage
		if err = rows.Scan(&raw); err != nil {
			return err
		}
		if err = enc.Encode(raw); err != nil {
			return err
		}
	}
	return rows.Err()
}
