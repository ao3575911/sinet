package api

import (
	"encoding/json"
)

func (c *call) one(q string, args ...any) (any, error) {
	var raw []byte
	err := c.tx.QueryRowContext(c.r.Context(), q, args...).Scan(&raw)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(raw), nil
}
func (c *call) many(q string, args ...any) (any, error) {
	rows, err := c.tx.QueryContext(c.r.Context(), q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []json.RawMessage{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		items = append(items, json.RawMessage(raw))
	}
	return map[string]any{"items": items}, rows.Err()
}
