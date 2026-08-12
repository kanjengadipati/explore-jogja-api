package ai

import "testing"

func TestUnmarshalJSONToleratesTrailingData(t *testing.T) {
	text := `{"name": "Prambanan", "ticket_price": "Rp50.000"}
  ]
}
`
	var out map[string]string
	if err := UnmarshalJSON(text, &out); err != nil {
		t.Fatalf("expected success with trailing garbage, got %v", err)
	}
	if out["name"] != "Prambanan" || out["ticket_price"] != "Rp50.000" {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestUnmarshalJSONRejectsTruncatedJSON(t *testing.T) {
	text := `{"name": "Prambanan", "ticket_price": "Rp50.000"`
	var out map[string]string
	if err := UnmarshalJSON(text, &out); err == nil {
		t.Fatal("expected error for truncated JSON, got nil")
	}
}
