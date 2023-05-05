package test

import (
	"encoding/json"
)

// helper, returns the jsonish value as string, or an error as string
// just to make tests easier to manage
// NOTE: we assume the error message is never a valid jsonish, so there is no ambiguity
func jsonish(v any) string {
	switch v := v.(type) {
	case json.RawMessage:
		return string(v)
	case []byte:
		if json.Valid(v) {
			return string(v)
		}
	}
	j, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}
	return string(j)
}
