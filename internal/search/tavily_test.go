package search

import (
	"strings"
	"testing"
)

func TestNewClientSearchDepthDefaults(t *testing.T) {
	cases := []struct {
		name  string
		value string
		want  string
	}{
		{"empty defaults to basic", "", "basic"},
		{"explicit basic", "basic", "basic"},
		{"advanced", "advanced", "advanced"},
		{"case-insensitive advanced", "ADVANCED", "advanced"},
		{"unknown falls back to basic", "deep", "basic"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewClient(Config{
				Enabled:     true,
				APIKey:      "k",
				SearchDepth: tc.value,
			})
			if c.searchDepth != tc.want {
				t.Fatalf("searchDepth = %q, want %q", c.searchDepth, tc.want)
			}
		})
	}
}

func TestNewClientIncludeFlags(t *testing.T) {
	c := NewClient(Config{
		Enabled:           true,
		APIKey:            "k",
		IncludeAnswer:     true,
		IncludeRawContent: true,
	})
	if !c.includeAnswer {
		t.Fatal("includeAnswer = false, want true")
	}
	if !c.includeRawContent {
		t.Fatal("includeRawContent = false, want true")
	}
}

func TestFormatResultsForPromptPrefersRawContent(t *testing.T) {
	out := FormatResultsForPrompt("q", []Result{{
		Title:      "T",
		Content:    "short snippet",
		RawContent: "full page text with the actual ticket price Rp50.000 and opening hours",
	}})
	if !strings.Contains(out, "full page text with the actual ticket price") {
		t.Fatalf("raw content not used as snippet:\n%s", out)
	}
	if strings.Contains(out, "short snippet") {
		t.Fatalf("plain content used over raw content:\n%s", out)
	}
}
