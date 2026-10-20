package connectors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"sinet/contracts"
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
	if len(raw) > 32768 || json.Unmarshal(raw, &value) != nil {
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
	if json.Unmarshal(args, &value) != nil {
		return errors.New("invalid arguments")
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
