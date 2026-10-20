package api

import (
	_ "embed"
	"html/template"
	"net/http"
	"net/url"
)

//go:embed wallet_link.html
var walletPageSource string
var walletPage = template.Must(template.New("wallet-link").Parse(walletPageSource))

func (s *Server) walletSigningPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.RawQuery != "" {
		output(w, 400, map[string]string{"error": "wallet_challenge_must_use_fragment"})
		return
	}
	origin, err := url.Parse(s.PublicURL)
	if err != nil || origin.Host == "" {
		output(w, 503, map[string]string{"error": "wallet_origin_unavailable"})
		return
	}
	nonce := random(16)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src 'nonce-"+nonce+"'; style-src 'unsafe-inline'; connect-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'")
	_ = walletPage.Execute(w, map[string]string{"Nonce": nonce, "Origin": s.PublicURL, "Domain": origin.Host})
}
