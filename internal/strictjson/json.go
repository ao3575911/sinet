// Package strictjson rejects ambiguous JSON at authorization boundaries.
package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Validate rejects duplicate object keys, invalid UTF-8, trailing values and
// excessive nesting. Callers retain their own request/response byte budgets.
func Validate(raw []byte) error {
	if !utf8.Valid(raw) {
		return errors.New("invalid JSON encoding")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	if err := value(d, 0); err != nil {
		return err
	}
	if _, err := d.Token(); err != io.EOF {
		return errors.New("trailing JSON data")
	}
	return nil
}
func value(d *json.Decoder, depth int) error {
	if depth > 64 {
		return errors.New("JSON nesting limit")
	}
	token, err := d.Token()
	if err != nil {
		return err
	}
	if number, ok := token.(json.Number); ok {
		text := number.String()
		if len(text) > 128 {
			return errors.New("JSON number size limit")
		}
		if i := strings.IndexAny(text, "eE"); i >= 0 {
			exponent, err := strconv.Atoi(text[i+1:])
			if err != nil || exponent < -308 || exponent > 308 {
				return errors.New("JSON number exponent limit")
			}
		}
	}
	switch token {
	case json.Delim('{'):
		seen := map[string]bool{}
		for d.More() {
			token, err := d.Token()
			if err != nil {
				return err
			}
			key, ok := token.(string)
			if !ok || seen[key] {
				return errors.New("duplicate or invalid JSON key")
			}
			seen[key] = true
			if err := value(d, depth+1); err != nil {
				return err
			}
		}
		token, err = d.Token()
		if err != nil || token != json.Delim('}') {
			return errors.New("invalid JSON object")
		}
	case json.Delim('['):
		for d.More() {
			if err := value(d, depth+1); err != nil {
				return err
			}
		}
		token, err = d.Token()
		if err != nil || token != json.Delim(']') {
			return errors.New("invalid JSON array")
		}
	case json.Delim('}'), json.Delim(']'):
		return errors.New("unexpected JSON delimiter")
	}
	return nil
}

// Decode preserves number literals rather than rounding them through float64.
func Decode(raw []byte, destination any) error {
	if err := Validate(raw); err != nil {
		return err
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	d.DisallowUnknownFields()
	return d.Decode(destination)
}
