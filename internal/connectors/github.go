package connectors

import (
	"encoding/json"
	"errors"
	"reflect"

	"sinet/contracts"
	"sinet/internal/githubpublic"
)

// githubContract reserves the reviewed ID for one fixed operation and closed inputs.
func githubContract(m Manifest) error {
	if m.ID != "github-public" {
		return nil
	}
	if m.Runtime != "remote_https" || len(m.Operations) != 1 {
		return errors.New("github-public requires the fixed HTTPS operation")
	}
	op := m.Operations[0]
	var actual, expected any
	_ = json.Unmarshal(op.Parameters, &actual)
	_ = json.Unmarshal([]byte(contracts.GitHubParameters), &expected)
	if op.Name != "repository_summary" || op.Risk != "external_disclosure" || !reflect.DeepEqual(actual, expected) {
		return errors.New("github-public contract must match the reviewed repository_summary schema")
	}
	return nil
}

// Disclosure is shared by CLI previews and server consent responses.
func Disclosure(m Manifest, op Operation, args json.RawMessage) map[string]any {
	recipients := []string{m.EndpointURL}
	purpose := op.Description
	if m.ID == "github-public" {
		recipients = append(recipients, "https://api.github.com")
		purpose = githubpublic.Disclosure
	}
	return map[string]any{"connector": m.ID, "version": m.Version, "operation": op.Name, "risk": op.Risk, "destination": m.EndpointURL, "recipients": recipients, "arguments": args, "disclosure": purpose}
}
