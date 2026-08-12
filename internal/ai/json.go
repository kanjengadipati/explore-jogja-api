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

// SanitizeJSONStrings escapes raw control characters (newlines, tabs, carriage
// returns) that Gemini occasionally emits INSIDE JSON string values. Go's
// encoding/json rejects raw control characters in strings, which made
// otherwise-valid model output fail to parse. It only rewrites characters
// inside string literals, leaving structural whitespace untouched.
func SanitizeJSONStrings(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	inString := false
	escaped := false
	for _, r := range s {
		if inString {
			if escaped {
				escaped = false
			} else {
				switch r {
				case '\\':
					escaped = true
				case '"':
					inString = false
				case '\n':
					b.WriteString(`\n`)
					continue
				case '\r':
					b.WriteString(`\r`)
					continue
				case '\t':
					b.WriteString(`\t`)
					continue
				}
			}
			b.WriteRune(r)
			continue
		}
		if r == '"' {
			inString = true
		}
		b.WriteRune(r)
	}
	return b.String()
}

// RepairUnterminatedJSON repairs model output that was truncated mid-structure:
//   - an unterminated string value (appends a closing quote),
//   - unclosed objects/arrays (appends "}"/"]" per open bracket, innermost
//     first, using a nesting stack so an open object inside an array closes
//     before the array does).
//
// Braces/brackets inside string literals are ignored. Valid JSON passes
// through unchanged.
func RepairUnterminatedJSON(s string) string {
	var stack []byte
	inString := false
	escaped := false
	for _, r := range s {
		if inString {
			if escaped {
				escaped = false
			} else if r == '\\' {
				escaped = true
			} else if r == '"' {
				inString = false
			}
			continue
		}
		switch r {
		case '"':
			inString = true
		case '{', '[':
			stack = append(stack, byte(r))
		case '}':
			if n := len(stack); n > 0 && stack[n-1] == '{' {
				stack = stack[:n-1]
			}
		case ']':
			if n := len(stack); n > 0 && stack[n-1] == '[' {
				stack = stack[:n-1]
			}
		}
	}
	var b strings.Builder
	if inString {
		b.WriteByte('"')
	}
	for i := len(stack) - 1; i >= 0; i-- {
		if stack[i] == '{' {
			b.WriteByte('}')
		} else {
			b.WriteByte(']')
		}
	}
	if b.Len() == 0 {
		return s
	}
	return s + b.String()
}
