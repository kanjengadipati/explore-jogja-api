package event

import (
	"encoding/json"
	"testing"
)

func TestParseCoord(t *testing.T) {
	str := func(s string) *string { return &s }
	cases := []struct {
		name    string
		input   *string
		want    float64
		wantErr bool
	}{
		{name: "nil", input: nil, want: 0},
		{name: "numeric", input: str("-7.7928"), want: -7.7928},
		{name: "empty string", input: str(""), want: 0},
		{name: "whitespace", input: str("   "), want: 0},
		{name: "invalid", input: str("abc"), wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseCoord(tc.input, "latitude")
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("parseCoord = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestEventUnmarshalJSONCoordinates(t *testing.T) {
	cases := []struct {
		name     string
		json     string
		wantLat  float64
		wantLng  float64
		wantErr  bool
	}{
		{name: "numeric", json: `{"id":"x","latitude":-7.7928,"longitude":110.3658}`, wantLat: -7.7928, wantLng: 110.3658},
		{name: "string", json: `{"id":"x","latitude":"-7.7928","longitude":"110.3658"}`, wantLat: -7.7928, wantLng: 110.3658},
		{name: "empty string", json: `{"id":"x","latitude":"","longitude":""}`, wantLat: 0, wantLng: 0},
		{name: "missing", json: `{"id":"x"}`, wantLat: 0, wantLng: 0},
		{name: "invalid string", json: `{"id":"x","latitude":"abc"}`, wantErr: true},
		{name: "invalid type", json: `{"id":"x","latitude":{"a":1}}`, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var e Event
			err := json.Unmarshal([]byte(tc.json), &e)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if e.Latitude != tc.wantLat {
				t.Errorf("latitude = %v, want %v", e.Latitude, tc.wantLat)
			}
			if e.Longitude != tc.wantLng {
				t.Errorf("longitude = %v, want %v", e.Longitude, tc.wantLng)
			}
			if e.ExternalID != "x" {
				t.Errorf("external_id = %q, want %q", e.ExternalID, "x")
			}
		})
	}
}
