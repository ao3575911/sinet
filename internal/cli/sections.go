package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"sinet/internal/gmail"
)

func (a *App) walletCommand(args []string) error {
	if len(args) == 0 || len(args) == 1 && args[0] == "list" {
		return a.do("GET", "/v1/wallets", nil)
	}
	if len(args) == 2 && args[0] == "challenge" {
		return a.do("POST", "/v1/wallets/challenge", map[string]string{"address": args[1]})
	}
	if len(args) == 2 && args[0] == "verify" {
		signature, err := a.In.ReadString('\n')
		if err != nil && err != io.EOF {
			return err
		}
		return a.do("POST", "/v1/wallets/verify", map[string]string{"nonce": args[1], "signature": strings.TrimSpace(signature)})
	}
	if len(args) == 3 && args[0] == "unlink" && args[2] == "--confirm-unlink" {
		return a.do("DELETE", "/v1/wallets/"+segment(args[1]), map[string]bool{"confirm": true})
	}
	if len(args) == 2 && args[0] == "connect" {
		challenge, err := a.C.Do("POST", "/v1/wallets/challenge", map[string]string{"address": args[1]})
		if err != nil {
			return err
		}
		message, ok := challenge["message"].(string)
		nonce, nonceOK := challenge["nonce"].(string)
		if !ok || !nonceOK {
			return errors.New("API returned an invalid wallet challenge")
		}
		_, _ = fmt.Fprintln(a.Out, "Sign this exact message in your existing Ethereum wallet using personal_sign. Keep the wallet's keys in the wallet. This links ownership and authorizes no transactions.")
		_, _ = fmt.Fprintln(a.Out, SafeText(message))
		if link, ok := challenge["signing_url"].(string); ok {
			_, _ = fmt.Fprintln(a.Out, "Browser wallet signing helper:", SafeText(link))
		}
		signature, err := a.line("Paste the one-time signature (0x…); empty cancels: ")
		if err != nil {
			return err
		}
		if signature == "" {
			return nil
		}
		return a.do("POST", "/v1/wallets/verify", map[string]string{"nonce": nonce, "signature": signature})
	}
	return errors.New("usage: wallet [list|challenge ADDRESS|verify NONCE|connect ADDRESS|unlink ADDRESS --confirm-unlink]")
}

func (a *App) emailCommand(args []string, query string) error {
	if len(args) == 0 || len(args) == 1 && args[0] == "status" {
		return a.do("GET", "/v1/email", nil)
	}
	if len(args) == 1 && args[0] == "inbox" {
		return a.print(map[string]any{"url": gmail.InboxURL, "mode": "external_gmail", "mailbox_synced": false})
	}
	if len(args) == 1 && args[0] == "list" {
		return a.do("GET", "/v1/email/drafts"+query, nil)
	}
	if len(args) == 2 && (args[0] == "get" || args[0] == "delete") {
		method := "GET"
		if args[0] == "delete" {
			method = "DELETE"
		}
		return a.do(method, "/v1/email/drafts/"+segment(args[1]), nil)
	}
	if len(args) == 3 && args[0] == "gmail" && args[2] == "--approve" {
		draft, err := a.C.Do("GET", "/v1/email/drafts/"+segment(args[1]), nil)
		if err != nil {
			return err
		}
		to, toOK := draft["recipient"].(string)
		subject, subjectOK := draft["subject"].(string)
		body, bodyOK := draft["body"].(string)
		if !toOK || !subjectOK || !bodyOK {
			return errors.New("API returned an invalid draft")
		}
		link, err := gmail.ComposeURL(to, subject, body)
		if err != nil {
			return err
		}
		return a.print(map[string]any{"url": link, "sent": false, "next_step": "Open the link in Gmail and review the sender and recipients before sending. Opening shares the draft with Google."})
	}
	if len(args) >= 4 && args[0] == "new" || len(args) >= 5 && args[0] == "edit" {
		method, path := "POST", "/v1/email/drafts"
		fields := args[1:]
		if args[0] == "edit" {
			method = "PATCH"
			path += "/" + segment(args[1])
			fields = args[2:]
		}
		body := strings.Join(fields[2:], " ")
		if body == "--stdin" {
			raw, err := io.ReadAll(io.LimitReader(a.In, 65537))
			if err != nil {
				return err
			}
			if len(raw) > 65536 {
				return errors.New("draft body exceeds 65536 bytes")
			}
			body = string(raw)
		}
		return a.do(method, path, map[string]string{"recipient": fields[0], "subject": fields[1], "body": body})
	}
	return errors.New("usage: email [status|inbox|list|get ID|delete ID|new TO SUBJECT BODY|edit ID TO SUBJECT BODY|gmail ID --approve]; opening a Gmail link shares recipient, subject and body with Google")
}

func (a *App) serviceMenu() error {
	choice, err := a.menu("Wallet, email and services", []string{"Wallet", "Email", "Business", "ENS and external identity", "Files and devices", "Terminal and cloud", "Mobile", "Help", "Back"})
	if err != nil {
		return err
	}
	switch choice {
	case 0:
		return a.walletMenu()
	case 1:
		return a.emailMenu()
	case 2, 3, 4, 5, 6:
		ids := map[int]string{2: "business", 3: "identity", 4: "storage", 5: "compute", 6: "mobile"}
		section, err := a.C.Do("GET", "/v1/sections/"+ids[choice], nil)
		if err != nil {
			return err
		}
		for _, key := range []string{"title", "status", "summary", "next_step"} {
			if text, ok := section[key].(string); ok {
				_, _ = fmt.Fprintln(a.Out, SafeText(text))
			}
		}
		return nil
	case 7:
		_, err = fmt.Fprint(a.Out, Help)
	}
	return err
}

func (a *App) walletMenu() error {
	if err := a.walletCommand(nil); err != nil {
		return err
	}
	choice, err := a.menu("Wallet", []string{"Link an Ethereum wallet", "Unlink a wallet", "Back"})
	if err != nil {
		return err
	}
	if choice == 2 {
		return nil
	}
	address, err := a.line("Ethereum address (0x…): ")
	if err != nil {
		return err
	}
	if choice == 0 {
		return a.walletCommand([]string{"connect", address})
	}
	confirm, err := a.line("Type unlink to remove this identity link; funds stay in your wallet: ")
	if err != nil {
		return err
	}
	if confirm != "unlink" {
		return nil
	}
	return a.walletCommand([]string{"unlink", address, "--confirm-unlink"})
}

func (a *App) emailMenu() error {
	if err := a.emailCommand(nil, ""); err != nil {
		return err
	}
	choice, err := a.menu("Email", []string{"Your drafts", "Compose a draft", "Review a draft in Gmail", "Open Gmail inbox link", "Delete a draft", "Back"})
	if err != nil {
		return err
	}
	switch choice {
	case 0:
		return a.emailCommand([]string{"list"}, "")
	case 1:
		to, err := a.line("Recipient (empty is allowed in a draft): ")
		if err != nil {
			return err
		}
		subject, err := a.line("Subject: ")
		if err != nil {
			return err
		}
		_, _ = fmt.Fprintln(a.Out, "Enter draft text. A single . saves; .cancel discards.")
		var body strings.Builder
		for {
			line, err := a.In.ReadString('\n')
			if err != nil {
				return err
			}
			if strings.TrimSpace(line) == ".cancel" {
				return nil
			}
			if strings.TrimSpace(line) == "." {
				break
			}
			body.WriteString(line)
			if body.Len() > 65536 {
				return errors.New("draft body exceeds 65536 bytes")
			}
		}
		return a.emailCommand([]string{"new", to, subject, body.String()}, "")
	case 2, 4:
		if err := a.emailCommand([]string{"list"}, ""); err != nil {
			return err
		}
		id, err := a.line("Draft ID: ")
		if err != nil {
			return err
		}
		if choice == 4 {
			return a.emailCommand([]string{"delete", id}, "")
		}
		confirm, err := a.line("Opening the Gmail link shares recipient, subject and body with Google. Type gmail to prepare it: ")
		if err != nil {
			return err
		}
		if confirm != "gmail" {
			return nil
		}
		return a.emailCommand([]string{"gmail", id, "--approve"}, "")
	case 3:
		return a.emailCommand([]string{"inbox"}, "")
	}
	return nil
}
