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

func TestRepairUnterminatedJSONRestoresMissingBraces(t *testing.T) {
	text := "{\"name\": \"Borobudur\", \"faqs\": [{\"q\": \"Harga?\", \"a\": \"Rp 60.000\"}]\n"
	repaired := RepairUnterminatedJSON(text)
	var out struct {
		Name string `json:"name"`
		Faqs []struct {
			Q string `json:"q"`
			A string `json:"a"`
		} `json:"faqs"`
	}
	if err := UnmarshalJSON(repaired, &out); err != nil {
		t.Fatalf("repaired output should parse, got %v (repaired: %q)", err, repaired)
	}
	if out.Name != "Borobudur" || len(out.Faqs) != 1 || out.Faqs[0].A != "Rp 60.000" {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestRepairUnterminatedJSONLeavesValidJSONUntouched(t *testing.T) {
	valid := `{"a": 1, "b": {"c": [1, 2]}}`
	if got := RepairUnterminatedJSON(valid); got != valid {
		t.Fatalf("valid JSON must be unchanged, got %q", got)
	}
}

func TestRepairUnterminatedJSONIgnoresBracesInsideStrings(t *testing.T) {
	truncated := `{"q": "Apa isi {kurung}?", "a": "Tidak apa"`
	repaired := RepairUnterminatedJSON(truncated)
	if repaired != truncated+"}" {
		t.Fatalf("expected exactly one closing brace appended, got %q", repaired)
	}
	var out map[string]string
	if err := UnmarshalJSON(repaired, &out); err != nil {
		t.Fatalf("repaired output should parse, got %v", err)
	}
	if out["q"] != "Apa isi {kurung}?" {
		t.Fatalf("unexpected value: %+v", out)
	}
}

func TestRepairUnterminatedJSONUnclosedArray(t *testing.T) {
	truncated := `{"name": "Borobudur", "faqs": [{"q": "Harga?", "a": "Rp 60.000"}`
	repaired := RepairUnterminatedJSON(truncated)
	var out struct {
		Name string `json:"name"`
		Faqs []struct {
			Q string `json:"q"`
			A string `json:"a"`
		} `json:"faqs"`
	}
	if err := UnmarshalJSON(repaired, &out); err != nil {
		t.Fatalf("repaired output should parse, got %v (repaired: %q)", err, repaired)
	}
	if out.Name != "Borobudur" || len(out.Faqs) != 1 || out.Faqs[0].A != "Rp 60.000" {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestRepairUnterminatedJSONUnterminatedString(t *testing.T) {
	truncated := `{"name": "Borobudur", "ticket_price": "Rp 60.000`
	repaired := RepairUnterminatedJSON(truncated)
	var out map[string]string
	if err := UnmarshalJSON(repaired, &out); err != nil {
		t.Fatalf("repaired output should parse, got %v (repaired: %q)", err, repaired)
	}
	if out["ticket_price"] != "Rp 60.000" {
		t.Fatalf("unexpected value: %+v", out)
	}
}

// TestRepairUnterminatedJSONRealCapture guards the exact truncated output a
// live gemini-3.5-flash call produced: it ended after the faqs array's "]"
// without closing the root object.
func TestRepairUnterminatedJSONRealCapture(t *testing.T) {
	text := `{
  "name": "Candi Borobudur",
  "name_en": "Borobudur Temple",
  "category": "Temple",
  "tagline": "Kemegahan Candi Buddha Terbesar di Dunia",
  "description": "Candi Borobudur merupakan salah satu situs warisan dunia UNESCO yang paling terkenal di dunia. Terletak di Kabupaten Magelang, Jawa Tengah, monumen megah ini memegang predikat sebagai candi Buddha terbesar di dunia sekaligus menjadi magnet wisata utama di kawasan Asia Tenggara. Struktur batunya yang tersusun rapi menyimpan ribuan panel relief yang menceritakan ajaran Buddha dan kehidupan masyarakat Jawa Kuno.\n\nBagi wisatawan yang ingin merasakan langsung atmosfer spiritual dan sejarahnya, komplek candi ini menyediakan akses untuk naik ke struktur utama maupun sekadar berjalan-jalan di area taman sekitarnya. Pengunjung lokal maupun mancanegara dapat mengeksplorasi setiap sudut pelataran batu yang megah ini sembari mempelajari nilai sejarah yang terkandung di dalamnya.",
  "facilities": [
    "Area Parkir",
    "Toko Souvenir UMKM",
    "Loket Tiket",
    "Taman Wisata",
    "Toilet"
  ],
  "travel_tips": [
    "Pesan tiket naik ke struktur candi jauh-jauh hari secara online karena kuota harian sangat terbatas.",
    "Gunakan tempat parkir paling ujung jika ingin berjalan lebih dekat menuju pintu masuk utama.",
    "Siapkan fisik yang prima karena jalur keluar akan mengarahkan Anda berjalan memutar melewati pasar suvenir UMKM.",
    "Datanglah sejak pagi hari agar terhindar dari terik matahari yang menyengat di siang hari."
  ],
  "faqs": [
    {
      "q": "Apakah wisatawan diperbolehkan naik ke atas struktur Candi Borobudur?",
      "a": "Ya, wisatawan diperbolehkan naik ke struktur candi dengan membeli tiket khusus seharga Rp 455.000 dan harus memesan slot waktu pendakian (pukul 08:30 hingga 15:30) terlebih dahulu."
    },
    {
      "q": "Pukul berapa area taman di sekitar Candi Borobudur mulai dibuka?",
      "a": "Area taman di sekeliling Candi Borobudur buka setiap hari mulai pukul 06:30 WIB dengan biaya masuk sebesar Rp 412.500."
    },
    {
      "q": "Berapa harga tiket masuk reguler Candi Borobudur untuk warga lokal?",
      "a": "Untuk warga negara Indonesia (WNI), harga tiket masuk reguler adalah Rp 60.000 untuk dewasa, Rp 30.000 untuk pelajar/mahasiswa, dan Rp 15.000 untuk anak-anak."
    }
  ]`
	var out struct {
		Name        string `json:"name"`
		TicketPrice string `json:"ticket_price"`
		Facilities  []string
		TravelTips  []string `json:"travel_tips"`
		Faqs        []struct {
			Q string `json:"q"`
			A string `json:"a"`
		} `json:"faqs"`
	}
	if err := UnmarshalJSON(RepairUnterminatedJSON(text), &out); err != nil {
		t.Fatalf("repaired capture should parse: %v", err)
	}
	if out.Name != "Candi Borobudur" || len(out.Facilities) != 5 || len(out.TravelTips) != 4 || len(out.Faqs) != 3 {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestRepairUnterminatedJSONStopsAtUnclosedRoot(t *testing.T) {
	// Model stopped mid-way: no root brace, unclosed faqs element.
	text := `{"name": "Pantai Indrayanti", "faqs": [{"q": "Harga?", "a": "Rp 10.000`
	repaired := RepairUnterminatedJSON(text)
	var out struct {
		Name string `json:"name"`
		Faqs []struct {
			Q string `json:"q"`
			A string `json:"a"`
		} `json:"faqs"`
	}
	if err := UnmarshalJSON(repaired, &out); err != nil {
		t.Fatalf("repaired output should parse, got %v (repaired: %q)", err, repaired)
	}
	if out.Name != "Pantai Indrayanti" || len(out.Faqs) != 1 || out.Faqs[0].A != "Rp 10.000" {
		t.Fatalf("unexpected result: %+v", out)
	}
}

func TestSanitizeJSONStringsEscapesRawControlChars(t *testing.T) {
	text := "{\"desc\": \"Line one\nLine two\", \"ok\": true}"
	cleaned := SanitizeJSONStrings(text)
	var out map[string]any
	if err := UnmarshalJSON(cleaned, &out); err != nil {
		t.Fatalf("cleaned output should parse, got %v (cleaned: %q)", err, cleaned)
	}
	if out["desc"] != "Line one\nLine two" {
		t.Fatalf("unexpected value: %v", out["desc"])
	}
	if got := SanitizeJSONStrings(`{"a": 1}`); got != `{"a": 1}` {
		t.Fatal("valid JSON should be unchanged")
	}
}
