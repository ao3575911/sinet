package strictjson

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestRejectAmbiguityAndLimits(t *testing.T) {
	for _, raw := range []string{`{"a":1,"a":2}`, `{"nested":{"a":1,"\u0061":2}}`, `{"a":[{"x":1,"x":2}]}`, `{} {}`, `1e-999999999`, `1e309`, strings.Repeat("9", 129), ``, `]`, "{\"x\":\"\xff\"}", strings.Repeat("[", 66) + "0" + strings.Repeat("]", 66)} {
		if err := Validate([]byte(raw)); err == nil {
			t.Fatalf("accepted ambiguous JSON %q", raw)
		}
	}
	var out map[string]any
	if err := Decode([]byte(`{"a":9007199254740993,"b":{"a":1}}`), &out); err != nil {
		t.Fatal(err)
	}
	if out["a"] != json.Number("9007199254740993") {
		t.Fatal("number rounded")
	}
}
func FuzzValidate(f *testing.F) {
	for _, seed := range []string{`{"a":1}`, `[null,true,false]`, `{"a":1,"a":2}`, `"text"`} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 131072 {
			t.Skip()
		}
		if Validate(raw) == nil && !json.Valid(raw) {
			t.Fatal("accepted invalid JSON")
		}
	})
}
