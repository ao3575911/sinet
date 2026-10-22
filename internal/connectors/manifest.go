package connectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"sinet/contracts"
	"sinet/internal/strictjson"
)

type Operation struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Risk        string          `json:"risk"`
	Parameters  json.RawMessage `json:"parameters"`
}
type Manifest struct {
	ID               string      `json:"id"`
	Version          string      `json:"version"`
	Publisher        string      `json:"publisher"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Runtime          string      `json:"runtime"`
	PrivacyPolicyURL string      `json:"privacy_policy_url,omitempty"`
	Operations       []Operation `json:"operations"`
	EndpointURL      string      `json:"endpoint_url,omitempty"`
}

func compile(raw string) (*jsonschema.Schema, error) {
	c := jsonschema.NewCompiler()
	c.AssertFormat = true
	c.LoadURL = func(string) (io.ReadCloser, error) { return nil, errors.New("external schema references denied") }
	if err := c.AddResource("https://schema.invalid/root.json", strings.NewReader(raw)); err != nil {
		return nil, err
	}
	return c.Compile("https://schema.invalid/root.json")
}

var manifestSchema = func() *jsonschema.Schema {
	s, err := compile(contracts.ConnectorSchema)
	if err != nil {
		panic(err)
	}
	return s
}()

func Parse(raw []byte) (Manifest, error) {
	var value any
	var m Manifest
	if len(raw) > 32768 || strictjson.Decode(raw, &value) != nil {
		return m, errors.New("invalid manifest JSON or size")
	}
	if err := manifestSchema.Validate(value); err != nil {
		return m, errors.New("manifest does not match connector schema")
	}
	if err := json.Unmarshal(raw, &m); err != nil {
		return m, err
	}
	seen := map[string]bool{}
	for _, op := range m.Operations {
		if seen[op.Name] {
			return m, errors.New("duplicate operation name")
		}
		seen[op.Name] = true
		if err := parameterBudget(op.Parameters); err != nil {
			return m, err
		}
		schema, err := compile(string(op.Parameters))
		if err != nil || schema == nil {
			return m, errors.New("invalid operation parameter schema or external reference")
		}
		var shape map[string]any
		_ = json.Unmarshal(op.Parameters, &shape)
		if shape["type"] != "object" || shape["additionalProperties"] != false {
			return m, errors.New("parameters must declare an object with additionalProperties false")
		}
	}
	if m.Runtime == "remote_https" {
		if _, err := Endpoint(m.EndpointURL); err != nil {
			return m, err
		}
		if m.PrivacyPolicyURL == "" {
			return m, errors.New("remote connector requires privacy_policy_url")
		}
	}
	if m.Runtime == "builtin" && m.EndpointURL != "" {
		return m, errors.New("builtin cannot have endpoint")
	}
	if err := githubContract(m); err != nil {
		return m, err
	}
	return m, nil
}
func (m Manifest) Operation(name string) (Operation, error) {
	for _, op := range m.Operations {
		if op.Name == name {
			return op, nil
		}
	}
	return Operation{}, errors.New("unknown operation")
}
func (op Operation) Validate(args json.RawMessage) error {
	if len(args) > 8192 {
		return errors.New("arguments too large")
	}
	var value any
	if strictjson.Decode(args, &value) != nil {
		return errors.New("invalid arguments")
	}
	if err := parameterBudget(op.Parameters); err != nil {
		return err
	}
	schema, err := compile(string(op.Parameters))
	if err != nil {
		return err
	}
	if schema.Validate(value) != nil {
		return fmt.Errorf("arguments do not match operation schema")
	}
	return nil
}

// Parameter schemas are untrusted at submission time, before operator review.
// Inline schemas bound evaluation work by the supplied tree size. Reference
// graphs can amplify a tiny document into exponential work without any I/O.
func parameterBudget(raw []byte) error {
	if len(raw) > 8192 {
		return errors.New("parameter schema size limit")
	}
	var value any
	if err := strictjson.Decode(raw, &value); err != nil {
		return errors.New("invalid parameter schema JSON")
	}
	nodes := 0
	var visit func(any, int) error
	visit = func(value any, depth int) error {
		nodes++
		if nodes > 512 || depth > 16 {
			return errors.New("parameter schema complexity limit")
		}
		switch value := value.(type) {
		case map[string]any:
			for key, child := range value {
				if key == "$ref" || key == "$recursiveRef" || key == "$dynamicRef" {
					return errors.New("parameter schemas must be inline; references denied")
				}
				if err := visit(child, depth+1); err != nil {
					return err
				}
			}
		case []any:
			for _, child := range value {
				if err := visit(child, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	return visit(value, 0)
}
