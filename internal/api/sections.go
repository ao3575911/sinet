package api

type section struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Status   string `json:"status"`
	Summary  string `json:"summary"`
	NextStep string `json:"next_step"`
}

var sections = []section{
	{"wallet", "Wallet", "available", "Link an Ethereum account by signing a one-time ownership message.", "Use your existing wallet to sign; keep its keys in your wallet."},
	{"email", "Email", "available", "Save private drafts and review them in your existing Gmail account.", "Open a draft in Gmail to review and send; mailbox sync needs Google OAuth setup."},
	{"business", "Business", "planned", "Business identities, team permissions and domain mail.", "Choose a business domain and define team ownership before setup."},
	{"identity", "ENS and external identity", "planned", "Optional external identifiers for your SINet identity.", "Wallet ownership linking is available; ENS resolution and DID verification are the next identity work."},
	{"storage", "Files and devices", "planned", "Device storage and file workflows alongside your private notes.", "Use Notes now; device access needs a separately installed local bridge."},
	{"compute", "Terminal and cloud", "planned", "Local terminal sessions and isolated cloud workspaces.", "A local bridge or isolated workspace provider must be set up first."},
	{"mobile", "Mobile", "planned", "A companion experience for your account on mobile.", "Use your public profile link now; a signed-in mobile interface is still to be built."},
}

func (c *call) sectionCatalog() (any, error) { return map[string]any{"items": sections}, nil }
func (c *call) sectionDetails() (any, error) {
	for _, item := range sections {
		if item.ID == c.r.PathValue("id") {
			return item, nil
		}
	}
	return nil, fail(404, "section_not_found")
}
