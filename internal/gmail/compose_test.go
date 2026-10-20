package gmail

import (
	"net/url"
	"strings"
	"testing"
)

func TestComposeHandoffIsFixedAndEncoded(t *testing.T) {
	to, subject, body := "User@example.test", "Hello &bcc=other@example.test", "Text\nhttps://evil.example/?x=1&z=2"
	link, err := ComposeURL(to, subject, body)
	if err != nil {
		t.Fatal(err)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	if u.Scheme != "https" || u.Host != "mail.google.com" || u.Path != "/mail/" || q.Get("to") != to || q.Get("su") != subject || q.Get("body") != body || q.Has("bcc") || len(q) != 5 {
		t.Fatal("handoff target or encoding changed", link)
	}
	for _, test := range []struct{ to, subject, body string }{
		{"x@example.test\r\nBcc: other@example.test", "subject", "body"},
		{"Name <x@example.test>", "subject", "body"},
		{to, "bad\nsubject", body},
		{to, "hidden\u202etext", body},
		{to, subject, strings.Repeat("x", 9000)},
	} {
		if _, err := ComposeURL(test.to, test.subject, test.body); err == nil {
			t.Fatal("unsafe or oversized handoff accepted")
		}
	}
}
