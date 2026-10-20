package connectors

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestManifestContractAndArguments(t *testing.T) {
	raw, err := os.ReadFile("../../connectors/examples/echo.json")
	if err != nil {
		t.Fatal(err)
	}
	m, err := Parse(raw)
	if err != nil {
		t.Fatal(err)
	}
	op, err := m.Operation("repeat")
	if err != nil {
		t.Fatal(err)
	}
	for _, args := range []string{`{"text":"hello"}`, `{"text":""}`} {
		if err = op.Validate(json.RawMessage(args)); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range []string{`{}`, `{"text":12}`, `{"text":"x","command":"cat /etc/passwd"}`, `{"text":"` + strings.Repeat("x", 513) + `"}`} {
		if op.Validate(json.RawMessage(args)) == nil {
			t.Fatal("invalid tool args accepted")
		}
	}
	var obj map[string]any
	_ = json.Unmarshal(raw, &obj)
	obj["command"] = "sh"
	b, _ := json.Marshal(obj)
	if _, err = Parse(b); err == nil {
		t.Fatal("executable metadata accepted")
	}
	delete(obj, "command")
	ops := obj["operations"].([]any)
	obj["operations"] = append(ops, ops[0])
	b, _ = json.Marshal(obj)
	if _, err = Parse(b); err == nil {
		t.Fatal("duplicate operations accepted")
	}
	obj["operations"] = ops
	ops[0].(map[string]any)["parameters"] = map[string]any{"$ref": "https://127.0.0.1/schema"}
	b, _ = json.Marshal(obj)
	if _, err = Parse(b); err == nil {
		t.Fatal("external schema reference accepted")
	}
}
