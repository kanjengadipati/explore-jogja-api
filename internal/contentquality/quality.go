// Package contentquality implements a lightweight style gate for AI-generated
// tourism copy. The system prompts already ask for a warm, conversational voice;
// this package enforces it by detecting stock brochure clichés in the generated
// draft and returning feedback that triggers a regeneration with a stronger hint.
package contentquality

import (
	"fmt"
	"strings"
)

// MaxAttempts bounds how many regeneration attempts a draft may use before the
// best available draft is accepted as-is.
const MaxAttempts = 3

// BannedPhrases are the stock tourism-brochure clichés that make AI copy read
// as templated regardless of the writing prompt. Each phrase is matched
// case-insensitively across the generated prose fields.
var BannedPhrases = []string{
	"menawarkan pengalaman tak terlupakan",
	"menawarkan pengalaman",
	"menawarkan keindahan",
	"pengalaman tak terlupakan",
	"pengalaman yang tak terlupakan",
	"tak terlupakan",
	"destinasi wisata yang menakjubkan",
	"tempat wisata yang menakjubkan",
	"wisata yang menakjubkan",
	"keindahan alam yang menakjubkan",
	"keindahan alam yang luar biasa",
	"keindahan yang luar biasa",
	"salah satu destinasi wisata",
	"salah satu tempat wisata",
	"destinasi wisata paling populer",
	"tempat wisata paling populer",
	"paling populer di yogyakarta",
	"unforgettable experience",
	"unforgettable journey",
	"unforgettable adventure",
	"once-in-a-lifetime",
	"breathtaking views",
	"hidden gem",
	"jaw-dropping",
}

// FindBanned returns the unique banned phrases present in text, lowercased
// during matching. An empty slice means the draft passed the style gate.
func FindBanned(text string) []string {
	lower := strings.ToLower(text)
	var found []string
	seen := make(map[string]struct{}, len(BannedPhrases))
	for _, p := range BannedPhrases {
		if _, ok := seen[p]; ok {
			continue
		}
		if strings.Contains(lower, p) {
			seen[p] = struct{}{}
			found = append(found, p)
		}
	}
	return found
}

// Feedback builds the instruction handed back to the model on a retry so the
// next draft avoids the flagged phrases without inventing new ones.
func Feedback(found []string) string {
	return fmt.Sprintf(
		"STYLE GATE FAILED: your previous draft used the clichéd tourism phrase(s): %s.\n"+
			"Rewrite the ENTIRE response from scratch WITHOUT those phrases. "+
			"Replace them with concrete sensory details, specific numbers, real facts, or local knowledge "+
			"instead of generic praise.",
		strings.Join(found, ", "),
	)
}

// Prose joins the prose-bearing fields of a generated payload so the style gate
// scans the copy that readers actually see, not just one field.
func Prose(parts ...string) string {
	nonEmpty := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			nonEmpty = append(nonEmpty, p)
		}
	}
	return strings.Join(nonEmpty, "\n")
}
