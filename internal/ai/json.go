package ai

import (
	"encoding/json"
	"strings"
)

// UnmarshalJSON parses the first complete JSON value in text into v, ignoring
// any trailing data. Some models append stray closing brackets/braces or notes
// after the actual JSON payload, which strict json.Unmarshal rejects with
// "extra data".
func UnmarshalJSON(text string, v any) error {
	dec := json.NewDecoder(strings.NewReader(text))
	return dec.Decode(v)
}
