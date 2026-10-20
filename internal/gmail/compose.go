// Package gmail prepares a user-reviewed Gmail web handoff without mailbox access.
package gmail

import (
	"errors"
	"net/mail"
	"net/url"
	"strings"
	"unicode"
)

const InboxURL = "https://mail.google.com/mail/u/0/#inbox"

// ComposeURL does not contact Google, send mail, or change draft state.
func ComposeURL(recipient, subject, body string) (string, error) {
	if recipient != "" {
		parsed, err := mail.ParseAddress(recipient)
		if err != nil || parsed.Address != recipient || len(recipient) > 254 {
			return "", errors.New("invalid recipient")
		}
	}
	if len(subject) > 200 || len(body) > 65536 {
		return "", errors.New("draft exceeds limits")
	}
	for _, value := range []string{recipient, subject} {
		if strings.ContainsFunc(value, func(r rune) bool { return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) }) {
			return "", errors.New("invalid draft header")
		}
	}
	values := url.Values{"view": {"cm"}, "fs": {"1"}, "to": {recipient}, "su": {subject}, "body": {body}}
	link := "https://mail.google.com/mail/?" + values.Encode()
	if len(link) > 8192 {
		return "", errors.New("draft too long for a Gmail link; shorten it or copy it into Gmail")
	}
	return link, nil
}
