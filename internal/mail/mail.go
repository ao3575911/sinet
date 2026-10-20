package mail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Sender interface {
	Send(context.Context, string, string) error
}
type SMTP struct {
	Addr, From, Username, Password string
	RequireTLS                     bool
}

func (s SMTP) Send(ctx context.Context, to, code string) error {
	if strings.ContainsAny(s.From+to, "\r\n") {
		return fmt.Errorf("invalid mail address")
	}
	host, _, err := net.SplitHostPort(s.Addr)
	if err != nil {
		return err
	}
	conn, err := (&net.Dialer{Timeout: 5 * time.Second}).DialContext(ctx, "tcp", s.Addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err = client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	} else if s.RequireTLS {
		return fmt.Errorf("SMTP STARTTLS required")
	}
	if s.Username != "" {
		if err = client.Auth(smtp.PlainAuth("", s.Username, s.Password, host)); err != nil {
			return err
		}
	}
	if err = client.Mail(s.From); err != nil {
		return err
	}
	if err = client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "From: %s\r\nTo: %s\r\nSubject: SINet sign-in code\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\nYour code is %s. It expires in 5 minutes and can be used once.\r\nNever share it with your agent or another person.\r\n", s.From, to, code)
	if err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return client.Quit()
}
