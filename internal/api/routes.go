package api

import "net/http"

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		if err := s.DB.PingContext(r.Context()); err != nil {
			output(w, 503, map[string]string{"status": "unavailable"})
			return
		}
		output(w, 200, map[string]string{"status": "ok"})
	})
	s.mux.HandleFunc("GET /v1/healthz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/healthz", http.StatusTemporaryRedirect)
	})
	s.mux.HandleFunc("GET /u/{handle}", s.webProfile)
	s.route("POST /v1/auth/request-otp", false, (*call).requestOTP)
	s.route("POST /v1/auth/verify-otp", false, (*call).verifyOTP)
	s.route("POST /v1/auth/logout", true, (*call).logout)
	s.route("GET /v1/me", true, (*call).me)
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
}
