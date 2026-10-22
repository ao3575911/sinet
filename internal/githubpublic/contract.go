// Package githubpublic defines the fixed reviewed public-repository contract.
package githubpublic

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const Input = `{"owner":"ao3575911","repository":"sinet"}`
const Upstream = "https://api.github.com/repos/ao3575911/sinet"
const PublicURL = "https://github.com/ao3575911/sinet"
const Disclosure = `Send owner "ao3575911" and repository "sinet" to the reviewed gateway and GitHub (api.github.com) to fetch public repository facts. No SINet session, email, notes, profile, model history or provider secrets are sent. No writes. Returned text is untrusted data; open-issue counts include pull requests.`

// Summary contains only explicitly approved public fields, never raw GitHub JSON.
type Summary struct {
	FullName    string  `json:"full_name"`
	HTMLURL     string  `json:"html_url"`
	Description *string `json:"description"`
	Language    *string `json:"language"`
	Stars       int64   `json:"stargazers_count"`
	Forks       int64   `json:"forks_count"`
	OpenIssues  int64   `json:"open_issues_count"`
	Archived    bool    `json:"archived"`
	UpdatedAt   string  `json:"updated_at"`
	Source      string  `json:"source"`
}

// object rejects duplicate keys and trailing values as well as malformed JSON.
// Raw fields allow upstream additions without ever forwarding them.
func object(data []byte) (map[string]json.RawMessage, error) {
	if !utf8.Valid(data) {
		return nil, errors.New("invalid UTF-8")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	token, err := decoder.Token()
	if err != nil || token != json.Delim('{') {
		return nil, errors.New("object required")
	}
	fields := make(map[string]json.RawMessage)
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return nil, err
		}
		key, ok := token.(string)
		if !ok {
			return nil, errors.New("invalid key")
		}
		if _, exists := fields[key]; exists {
			return nil, errors.New("duplicate key")
		}
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			return nil, err
		}
		fields[key] = raw
	}
	if _, err := decoder.Token(); err != nil {
		return nil, err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return nil, errors.New("trailing data")
	}
	return fields, nil
}

func ValidInput(data []byte) bool {
	if len(data) > 512 {
		return false
	}
	fields, err := object(data)
	if err != nil || len(fields) != 2 {
		return false
	}
	var owner, repository string
	return json.Unmarshal(fields["owner"], &owner) == nil &&
		json.Unmarshal(fields["repository"], &repository) == nil &&
		owner == "ao3575911" && repository == "sinet"
}

func safeText(text string, limit int) bool {
	if len(text) > limit {
		return false
	}
	for _, r := range text {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == unicode.ReplacementChar {
			return false
		}
	}
	return true
}

func ParseUpstream(data []byte) (Summary, error) {
	var summary Summary
	fields, err := object(data)
	if err != nil {
		return summary, err
	}
	var private bool
	values := map[string]any{
		"full_name": &summary.FullName, "html_url": &summary.HTMLURL,
		"description": &summary.Description, "language": &summary.Language,
		"stargazers_count": &summary.Stars, "forks_count": &summary.Forks,
		"open_issues_count": &summary.OpenIssues, "archived": &summary.Archived,
		"updated_at": &summary.UpdatedAt, "private": &private,
	}
	for name, destination := range values {
		raw, exists := fields[name]
		if !exists || (name != "description" && name != "language" && strings.TrimSpace(string(raw)) == "null") {
			return summary, errors.New("missing upstream field")
		}
		if err := json.Unmarshal(raw, destination); err != nil {
			return summary, err
		}
	}
	if private || summary.FullName != "ao3575911/sinet" || summary.HTMLURL != PublicURL {
		return summary, errors.New("unexpected repository")
	}
	for _, count := range []int64{summary.Stars, summary.Forks, summary.OpenIssues} {
		if count < 0 || count > 9007199254740991 {
			return summary, errors.New("invalid count")
		}
	}
	if summary.Description != nil && !safeText(*summary.Description, 1024) {
		return summary, errors.New("invalid description")
	}
	if summary.Language != nil && !safeText(*summary.Language, 128) {
		return summary, errors.New("invalid language")
	}
	if len(summary.UpdatedAt) > 40 {
		return summary, errors.New("invalid timestamp")
	}
	if _, err := time.Parse(time.RFC3339, summary.UpdatedAt); err != nil {
		return summary, err
	}
	summary.Source = Upstream
	return summary, nil
}

// ParseOutput revalidates a gateway response before the API audits success or renders it.
// It rejects unknown/duplicate fields; only raw GitHub responses may contain additions.
func ParseOutput(data []byte) (Summary, error) {
	var empty Summary
	if len(data) > 8192 {
		return empty, errors.New("summary output limit")
	}
	fields, err := object(data)
	if err != nil || len(fields) != 10 {
		return empty, errors.New("invalid summary fields")
	}
	var source string
	if json.Unmarshal(fields["source"], &source) != nil || source != Upstream {
		return empty, errors.New("invalid summary source")
	}
	delete(fields, "source")
	fields["private"] = json.RawMessage(`false`)
	// ParseUpstream requires every other field and validates its exact type and byte budget.
	normalized, err := json.Marshal(fields)
	if err != nil {
		return empty, err
	}
	return ParseUpstream(normalized)
}
