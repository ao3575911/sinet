package api

import (
	"context"
	"net/http"
	"time"
)

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := s.DB.PingContext(ctx); err != nil {
			output(w, 503, map[string]string{"status": "unavailable"})
			return
		}
		output(w, 200, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("GET /v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/healthz", http.StatusTemporaryRedirect)
	})
	s.mux.HandleFunc("GET /u/{handle}", s.webProfile)
	s.mux.HandleFunc("GET /wallet/link", s.walletSigningPage)
	s.route("POST /v1/auth/request-otp", false, (*call).requestOTP)
	s.route("POST /v1/auth/verify-otp", false, (*call).verifyOTP)
	s.route("POST /v1/auth/logout", true, (*call).logout)
	s.route("GET /v1/me", true, (*call).me)
	s.route("GET /v1/sections", true, (*call).sectionCatalog)
	s.route("GET /v1/sections/{id}", true, (*call).sectionDetails)
	s.route("GET /v1/wallets", true, (*call).wallets)
	s.route("POST /v1/wallets/challenge", true, (*call).walletChallenge)
	s.route("POST /v1/wallets/verify", true, (*call).verifyWallet)
	s.route("DELETE /v1/wallets/{address}", true, (*call).unlinkWallet)
	s.route("GET /v1/email", true, (*call).emailStatus)
	s.route("GET /v1/email/drafts", true, (*call).emailDrafts)
	s.route("POST /v1/email/drafts", true, (*call).saveEmailDraft)
	s.route("GET /v1/email/drafts/{id}", true, (*call).emailDraft)
	s.route("PATCH /v1/email/drafts/{id}", true, (*call).saveEmailDraft)
	s.route("DELETE /v1/email/drafts/{id}", true, (*call).deleteEmailDraft)
	s.route("PATCH /v1/me", true, (*call).patchMe)
	s.route("GET /v1/me/export", true, (*call).export)
	s.route("POST /v1/me/delete", true, (*call).deleteAccount)
	s.route("GET /v1/handles/{name}/availability", true, (*call).availability)
	s.route("POST /v1/handles/claim", true, (*call).claim)
	s.route("GET /v1/users/{handle}", false, (*call).publicProfile)
	s.route("GET /v1/users", true, (*call).users)
	s.route("GET /v1/rooms", true, (*call).rooms)
	s.route("POST /v1/rooms/{slug}/join", true, (*call).join)
	s.route("GET /v1/rooms/{slug}/messages", true, (*call).messages)
	s.route("POST /v1/rooms/{slug}/messages", true, (*call).postMessage)
	s.route("GET /v1/blocks", true, (*call).blocks)
	s.route("POST /v1/blocks", true, (*call).block)
	s.route("DELETE /v1/blocks/{handle}", true, (*call).unblock)
	s.route("POST /v1/reports", true, (*call).report)
	s.route("GET /v1/notes", true, (*call).notes)
	s.route("POST /v1/notes", true, (*call).saveNote)
	s.route("GET /v1/notes/{id}", true, (*call).note)
	s.route("PATCH /v1/notes/{id}", true, (*call).saveNote)
	s.route("DELETE /v1/notes/{id}", true, (*call).deleteNote)

	s.route("GET /v1/connectors", true, (*call).connectors)
	s.route("GET /v1/connectors/installed", true, (*call).installs)
	s.route("GET /v1/connectors/submissions", true, (*call).submissions)
	s.route("POST /v1/connectors/submissions", true, (*call).submitConnector)
	s.route("GET /v1/connectors/{id}", true, (*call).connector)
	s.route("POST /v1/connectors/{id}/install", true, (*call).install)
	s.route("DELETE /v1/connectors/{id}/install", true, (*call).uninstall)
	s.route("POST /v1/connectors/{id}/grants", true, (*call).grant)
	s.route("DELETE /v1/connectors/{id}/grants/{operation}", true, (*call).revoke)
	s.route("POST /v1/agent/consents", true, (*call).consent)
	s.route("POST /v1/agent/messages", true, (*call).agentMessage)
	s.route("GET /v1/agent/messages", true, (*call).agentHistory)
	s.route("GET /v1/audit/tool-calls", true, (*call).toolAudit)
	s.route("GET /v1/usage", true, (*call).usage)
}
