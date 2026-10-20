package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/term"
	"sinet/internal/client"
	"sinet/internal/connectors"
)

const Help = `SINet · Your terminal. Your network.
Usage: SINet [command]
No command opens guided onboarding and a keyboard dashboard in an interactive terminal.
All responses are JSON except help/share cards. Errors go to stderr, exit code 1.

  version | --version                 Print local semantic version
  health                              Check API health
  login                               Guided email and code sign-in
  login request EMAIL                 Send OTP (development: view in Mailpit)
  login verify EMAIL                  Read OTP from stdin; save private session
  logout                              Revoke this session and remove local credential
  claim HANDLE                        Claim a normalized unique handle
  me | user HANDLE | find QUERY       Own profile, public profile, directory
  share                               Shareable text profile card
  rooms                               List community rooms
  room SLUG join|read|say [MESSAGE]    Join, read, post (join required)
  note list | get ID                  Private persistent notes
  note new TITLE BODY                 Save a note; --stdin in place of BODY reads stdin
  note edit ID TITLE BODY             Replace a note; --stdin reads stdin
  note delete ID                      Delete own note
  plug list | installed | submissions Browse registry or own state
  plug info ID [VERSION]              Inspect manifest, provenance and capabilities
  plug validate FILE | submit FILE    Explicitly read a local manifest; submission enters review
  plug install ID VERSION             Install approved immutable version
  plug remove ID                      Remove installation and grants
  plug grant ID VERSION OP --approve  Explicit per-operation grant
  plug revoke ID OP                   Revoke operation and pending consent
  agent history | audit | usage       Own chat, tool audit and usage
  agent ask MESSAGE [--allow-model]   Explicit provider disclosure if AI is configured
  agent call ID VERSION OP JSON [--approve]
                                     Invoke tool; --approve consents to this exact call
  settings                            Current settings
  settings set KEY VALUE              theme, public, notifications, display_name, bio
  block HANDLE | unblock HANDLE       Filter directory/room interaction in both directions
  report user HANDLE REASON           Submit abuse report
  report message ID REASON            Report a message in a joined room
  export                              Own account data as JSON
  delete-account --confirm-delete     Irreversible account deletion

List pages: append --limit N (1–50), --offset N for catalog/notes/directory/blocks,
or --before ID for messages/history/audit. Repeat using the last returned ID.
API: SINET_API (HTTPS, or loopback HTTP). Session: SINET_CONFIG.
No host shell, USB, arbitrary plugin execution, active inbox or wallet is provided.
`

type App struct {
	C     *client.Client
	In    *bufio.Reader
	Out   io.Writer
	TTY   bool
	Input *os.File
}

func Run(c *client.Client, args []string, input *os.File, out io.Writer) error {
	app := &App{C: c, In: bufio.NewReader(input), Out: out, TTY: term.IsTerminal(int(input.Fd())), Input: input}
	if len(args) == 0 {
		if app.TTY {
			return app.dashboard()
		}
		_, err := fmt.Fprint(out, Help)
		return err
	}
	return app.run(args)
}
func SafeText(s string) string {
	return strings.Map(func(r rune) rune {
		if (unicode.IsControl(r) && r != '\n' && r != '\t') || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
}
func (a *App) print(v any) error {
	enc := json.NewEncoder(a.Out)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
func (a *App) do(method, path string, body any) error {
	out, err := a.C.Do(method, path, body)
	if err != nil {
		return err
	}
	return a.print(out)
}
func (a *App) line(prompt string) (string, error) {
	_, _ = fmt.Fprint(a.Out, prompt)
	line, err := a.In.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
func (a *App) otp() (string, error) {
	if a.TTY {
		_, _ = fmt.Fprint(a.Out, "Sign-in code (hidden): ")
		raw, err := term.ReadPassword(int(a.Input.Fd()))
		_, _ = fmt.Fprintln(a.Out)
		return strings.TrimSpace(string(raw)), err
	}
	line, err := a.In.ReadString('\n')
	if err != nil && err != io.EOF {
		return "", err
	}
	return strings.TrimSpace(line), nil
}
func (a *App) login(email string) error {
	if email == "" {
		var err error
		email, err = a.line("Email: ")
		if err != nil {
			return err
		}
	}
	if _, err := a.C.Do("POST", "/v1/auth/request-otp", map[string]string{"email": email}); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(a.Out, "Code requested. Check your email (local development: Mailpit at http://127.0.0.1:8025).")
	return a.verify(email)
}
func (a *App) verify(email string) error {
	code, err := a.otp()
	if err != nil {
		return err
	}
	out, err := a.C.Do("POST", "/v1/auth/verify-otp", map[string]string{"email": email, "code": code})
	if err != nil {
		return err
	}
	token, ok := out["token"].(string)
	if !ok {
		return errors.New("API did not issue a session")
	}
	if err = a.C.Save(token); err != nil {
		return err
	}
	return a.print(map[string]any{"signed_in": true, "expires_in": out["expires_in"]})
}
func segment(v string) string { return url.PathEscape(v) }
func (a *App) run(args []string) error {
	if len(args) == 0 {
		return nil
	}
	command := args[0]
	rest := args[1:]
	pathQuery := ""
	var queryErr error
	if command != "agent" && command != "login" && command != "note" || command == "note" && len(rest) > 0 && rest[0] == "list" {
		rest, pathQuery, queryErr = pageFlags(rest)
		if queryErr != nil {
			return queryErr
		}
	}
	switch command {
	case "help", "--help", "-h":
		_, err := fmt.Fprint(a.Out, Help)
		return err
	case "health":
		return a.do("GET", "/healthz", nil)
	case "login":
		if len(rest) == 0 {
			return a.login("")
		}
		if len(rest) != 2 {
			return errors.New("usage: login request|verify EMAIL")
		}
		if rest[0] == "request" {
			return a.do("POST", "/v1/auth/request-otp", map[string]string{"email": rest[1]})
		}
		if rest[0] == "verify" {
			return a.verify(rest[1])
		}
	case "logout":
		if _, err := a.C.Do("POST", "/v1/auth/logout", map[string]any{}); err != nil {
			return err
		}
		if err := a.C.Forget(); err != nil {
			return err
		}
		return a.print(map[string]bool{"logged_out": true})
	case "me", "settings":
		if len(rest) == 0 {
			return a.do("GET", "/v1/me", nil)
		}
		if command == "settings" && len(rest) >= 3 && rest[0] == "set" {
			key, value := rest[1], strings.Join(rest[2:], " ")
			var v any = value
			switch key {
			case "theme", "display_name", "bio":
			case "public", "notifications":
				b, err := strconv.ParseBool(value)
				if err != nil {
					return err
				}
				v = b
			default:
				return errors.New("unknown setting")
			}
			return a.do("PATCH", "/v1/me", map[string]any{key: v})
		}
	case "claim":
		if len(rest) == 1 {
			return a.do("POST", "/v1/handles/claim", map[string]string{"handle": rest[0]})
		}
	case "user":
		if len(rest) == 1 {
			return a.do("GET", "/v1/users/"+segment(rest[0]), nil)
		}
	case "find":
		if len(rest) > 0 {
			values := url.Values{"q": []string{strings.Join(rest, " ")}}
			if pathQuery != "" {
				page, _ := url.ParseQuery(strings.TrimPrefix(pathQuery, "?"))
				for k, v := range page {
					values[k] = v
				}
			}
			return a.do("GET", "/v1/users?"+values.Encode(), nil)
		}
	case "share":
		me, err := a.C.Do("GET", "/v1/me", nil)
		if err != nil {
			return err
		}
		handle, _ := me["handle"].(string)
		if handle == "" {
			return errors.New("claim a handle first")
		}
		_, err = fmt.Fprintf(a.Out, "@%s · SINet\n%s\n%s\nSINet user %s\n", SafeText(handle), SafeText(fmt.Sprint(me["bio"])), SafeText(fmt.Sprint(me["profile_url"])), SafeText(handle))
		return err
	case "rooms":
		return a.do("GET", "/v1/rooms"+pathQuery, nil)
	case "room":
		if len(rest) == 1 {
			if err := a.do("POST", "/v1/rooms/"+segment(rest[0])+"/join", map[string]any{}); err != nil {
				return err
			}
			return a.do("GET", "/v1/rooms/"+segment(rest[0])+"/messages"+pathQuery, nil)
		}
		if len(rest) >= 2 {
			base := "/v1/rooms/" + segment(rest[0])
			switch rest[1] {
			case "join":
				return a.do("POST", base+"/join", map[string]any{})
			case "read":
				return a.do("GET", base+"/messages"+pathQuery, nil)
			case "say":
				if len(rest) > 2 {
					return a.do("POST", base+"/messages", map[string]string{"content": strings.Join(rest[2:], " ")})
				}
			}
		}
	case "note":
		return a.note(rest, pathQuery)
	case "plug":
		return a.plug(rest, pathQuery)
	case "agent":
		return a.agent(rest)
	case "block":
		if len(rest) == 1 {
			return a.do("POST", "/v1/blocks", map[string]string{"handle": rest[0]})
		}
	case "unblock":
		if len(rest) == 1 {
			return a.do("DELETE", "/v1/blocks/"+segment(rest[0]), nil)
		}
	case "blocks":
		return a.do("GET", "/v1/blocks"+pathQuery, nil)
	case "report":
		if len(rest) >= 3 {
			body := map[string]any{"reason": strings.Join(rest[2:], " ")}
			if rest[0] == "user" {
				body["handle"] = rest[1]
			} else if rest[0] == "message" {
				id, err := strconv.ParseInt(rest[1], 10, 64)
				if err != nil {
					return err
				}
				body["message_id"] = id
			} else {
				return errors.New("report user|message TARGET REASON")
			}
			return a.do("POST", "/v1/reports", body)
		}
	case "export":
		return a.do("GET", "/v1/me/export"+pathQuery, nil)
	case "delete-account":
		if len(rest) == 1 && rest[0] == "--confirm-delete" {
			if _, err := a.C.Do("POST", "/v1/me/delete", map[string]string{"confirm": "delete my account"}); err != nil {
				return err
			}
			if err := a.C.Forget(); err != nil {
				return err
			}
			return a.print(map[string]bool{"deleted": true})
		}
		return errors.New("use delete-account --confirm-delete to explicitly authorize deletion")
	}
	return errors.New("invalid command or arguments; run SINet help")
}
func pageFlags(args []string) ([]string, string, error) {
	rest := []string{}
	q := url.Values{}
	for i := 0; i < len(args); i++ {
		if args[i] == "--limit" || args[i] == "--offset" || args[i] == "--before" {
			if i+1 == len(args) {
				return nil, "", errors.New("pagination flag needs a value")
			}
			n, err := strconv.ParseInt(args[i+1], 10, 64)
			if err != nil || n < 0 {
				return nil, "", errors.New("invalid pagination value")
			}
			q.Set(strings.TrimPrefix(args[i], "--"), args[i+1])
			i++
		} else {
			rest = append(rest, args[i])
		}
	}
	if len(q) == 0 {
		return rest, "", nil
	}
	return rest, "?" + q.Encode(), nil
}
func (a *App) body(args []string) (string, error) {
	if len(args) == 1 && args[0] == "--stdin" {
		raw, err := io.ReadAll(io.LimitReader(a.In, 65537))
		if err != nil {
			return "", err
		}
		if len(raw) > 65536 {
			return "", errors.New("note input exceeds 65536 bytes")
		}
		return string(raw), nil
	}
	return strings.Join(args, " "), nil
}
func (a *App) note(args []string, q string) error {
	if len(args) == 0 {
		if a.TTY {
			return a.noteEditor()
		}
		return errors.New("use note list|new|edit|get|delete")
	}
	switch args[0] {
	case "list":
		return a.do("GET", "/v1/notes"+q, nil)
	case "get":
		if len(args) == 2 {
			return a.do("GET", "/v1/notes/"+segment(args[1]), nil)
		}
	case "delete":
		if len(args) == 2 {
			return a.do("DELETE", "/v1/notes/"+segment(args[1]), nil)
		}
	case "new":
		if len(args) >= 3 {
			body, err := a.body(args[2:])
			if err != nil {
				return err
			}
			return a.do("POST", "/v1/notes", map[string]string{"title": args[1], "body": body})
		}
	case "edit":
		if len(args) >= 4 {
			body, err := a.body(args[3:])
			if err != nil {
				return err
			}
			return a.do("PATCH", "/v1/notes/"+segment(args[1]), map[string]string{"title": args[2], "body": body})
		}
	}
	return errors.New("invalid note command; see help")
}
func (a *App) plug(args []string, q string) error {
	if len(args) == 0 {
		return errors.New("use plug list|info|install|grant|revoke|remove|submit|validate|installed|submissions")
	}
	switch args[0] {
	case "list":
		return a.do("GET", "/v1/connectors"+q, nil)
	case "installed":
		return a.do("GET", "/v1/connectors/installed"+q, nil)
	case "submissions":
		return a.do("GET", "/v1/connectors/submissions"+q, nil)
	case "info":
		if len(args) == 2 || len(args) == 3 {
			path := "/v1/connectors/" + segment(args[1])
			if len(args) == 3 {
				path += "?version=" + url.QueryEscape(args[2])
			}
			return a.do("GET", path, nil)
		}
	case "install":
		if len(args) == 3 {
			return a.do("POST", "/v1/connectors/"+segment(args[1])+"/install", map[string]string{"version": args[2]})
		}
	case "remove":
		if len(args) == 2 {
			return a.do("DELETE", "/v1/connectors/"+segment(args[1])+"/install", nil)
		}
	case "grant":
		if len(args) == 5 && args[4] == "--approve" {
			return a.do("POST", "/v1/connectors/"+segment(args[1])+"/grants", map[string]any{"version": args[2], "operation": args[3], "approve": true})
		}
		return errors.New("inspect plug info ID VERSION, then plug grant ID VERSION OP --approve")
	case "revoke":
		if len(args) == 3 {
			return a.do("DELETE", "/v1/connectors/"+segment(args[1])+"/grants/"+segment(args[2]), nil)
		}
	case "validate", "submit":
		if len(args) == 2 {
			f, err := os.Open(args[1])
			if err != nil {
				return err
			}
			defer f.Close()
			raw, err := io.ReadAll(io.LimitReader(f, 32769))
			if err != nil {
				return err
			}
			if _, err = connectors.Parse(raw); err != nil {
				return err
			}
			if args[0] == "validate" {
				return a.print(map[string]bool{"valid": true})
			}
			return a.do("POST", "/v1/connectors/submissions", json.RawMessage(raw))
		}
	}
	return errors.New("invalid plug command; see help")
}
func (a *App) agent(args []string) error {
	if len(args) == 0 {
		return errors.New("use agent ask|call|history|audit|usage")
	}
	switch args[0] {
	case "history", "audit", "usage":
		rest, q, err := pageFlags(args[1:])
		if err != nil {
			return err
		}
		if len(rest) > 0 {
			return errors.New("unexpected arguments")
		}
		path := "/v1/agent/messages"
		if args[0] == "audit" {
			path = "/v1/audit/tool-calls"
		}
		if args[0] == "usage" {
			path = "/v1/usage"
		}
		return a.do("GET", path+q, nil)
	case "ask":
		if len(args) >= 2 {
			allow := args[len(args)-1] == "--allow-model"
			if allow {
				args = args[:len(args)-1]
			}
			return a.do("POST", "/v1/agent/messages", map[string]any{"message": strings.Join(args[1:], " "), "allow_model": allow})
		}
	case "call":
		if len(args) != 5 && len(args) != 6 {
			return errors.New("usage: agent call ID VERSION OP JSON [--approve]")
		}
		approve := len(args) == 6 && args[5] == "--approve"
		if len(args) == 6 && !approve {
			return errors.New("unknown call flag")
		}
		var arguments map[string]any
		if json.Unmarshal([]byte(args[4]), &arguments) != nil {
			return errors.New("arguments must be a JSON object")
		}
		info, err := a.C.Do("GET", "/v1/connectors/"+segment(args[1])+"?version="+url.QueryEscape(args[2]), nil)
		if err != nil {
			return err
		}
		raw, _ := json.Marshal(info["manifest"])
		manifest, err := connectors.Parse(raw)
		if err != nil {
			return err
		}
		op, err := manifest.Operation(args[3])
		if err != nil {
			return err
		}
		if err = op.Validate(json.RawMessage(args[4])); err != nil {
			return err
		}
		tool := map[string]any{"connector": args[1], "version": args[2], "operation": args[3], "arguments": arguments}
		if manifest.Runtime == "remote_https" || op.Risk != "read" {
			if !approve {
				if err = a.print(map[string]any{"approval_required": true, "destination": manifest.EndpointURL, "risk": op.Risk, "operation": op.Name, "arguments": arguments}); err != nil {
					return err
				}
				return errors.New("review destination and data; repeat with --approve to authorize this exact call")
			}
			consent, err := a.C.Do("POST", "/v1/agent/consents", map[string]any{"tool": tool, "approve": true})
			if err != nil {
				return err
			}
			tool["consent"] = consent["consent"]
		}
		return a.do("POST", "/v1/agent/messages", map[string]any{"tool": tool})
	}
	return errors.New("invalid agent command; see help")
}
