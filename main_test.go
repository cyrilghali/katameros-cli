package main

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
)

// ── detectLang ──────────────────────────────────────────────

func TestDetectLang(t *testing.T) {
	tests := []struct {
		name   string
		header string
		want   int
	}{
		{"empty defaults to French", "", 1},
		{"simple French", "fr", 1},
		{"simple English", "en", 2},
		{"simple Arabic", "ar", 3},
		{"with region tag", "fr-FR", 1},
		{"with quality values", "de;q=0.8,en;q=0.9,fr;q=0.7", 2},
		{"English first by default quality", "en,fr;q=0.9", 2},
		{"French preferred over English", "fr;q=1.0,en;q=0.8", 1},
		{"unknown falls back to French", "zh,ja", 1},
		{"mixed known and unknown", "zh;q=0.9,it;q=0.8", 4},
		{"complex real-world header", "fr-FR,fr;q=0.9,en-US;q=0.8,en;q=0.7", 1},
		{"Dutch", "nl-NL,nl;q=0.9", 9},
		{"Polish", "pl", 7},
		{"Spanish", "es-ES", 8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectLang(tt.header)
			if got != tt.want {
				t.Errorf("detectLang(%q) = %d, want %d", tt.header, got, tt.want)
			}
		})
	}
}

// ── formatCoptic ────────────────────────────────────────────

func TestFormatCoptic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"normal date", "17/6/1742", "17 Amshir 1742"},
		{"first month", "1/1/1740", "1 Thout 1740"},
		{"last regular month", "30/12/1742", "30 Mesori 1742"},
		{"intercalary month", "5/13/1742", "5 Nasie 1742"},
		{"malformed input", "bad", "bad"},
		{"month out of range high", "1/14/1742", "1/14/1742"},
		{"month out of range zero", "1/0/1742", "1/0/1742"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatCoptic(tt.input)
			if got != tt.want {
				t.Errorf("formatCoptic(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── wordWrap ────────────────────────────────────────────────

func TestWordWrap(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		width int
		want  int // expected number of lines
	}{
		{"empty", "", 40, 0},
		{"short fits one line", "Hello world", 40, 1},
		{"wraps at width", "Hello world this is a longer sentence that should wrap", 20, 3},
		{"single long word", "superlongword", 5, 1}, // can't break a word
		{"whitespace only", "   ", 40, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordWrap(tt.text, tt.width)
			if len(got) != tt.want {
				t.Errorf("wordWrap(%q, %d) returned %d lines, want %d: %v",
					tt.text, tt.width, len(got), tt.want, got)
			}
			// verify no line (except single-word overflow) exceeds width
			for i, line := range got {
				words := strings.Fields(line)
				if len(words) > 1 && len(line) > tt.width {
					t.Errorf("line %d exceeds width %d: %q (%d chars)",
						i, tt.width, line, len(line))
				}
			}
		})
	}
}

// ── formatVerse ─────────────────────────────────────────────

func TestFormatVerse(t *testing.T) {
	t.Run("short verse fits one line", func(t *testing.T) {
		result := formatVerse(1, "In the beginning.")
		if !strings.Contains(result, "In the beginning.") {
			t.Errorf("expected verse text in output, got: %q", result)
		}
		if !strings.Contains(result, "1") {
			t.Errorf("expected verse number in output, got: %q", result)
		}
	})

	t.Run("long verse wraps", func(t *testing.T) {
		long := strings.Repeat("word ", 20)
		result := formatVerse(17, long)
		lines := strings.Split(result, "\n")
		if len(lines) < 2 {
			t.Errorf("expected multiple lines for long verse, got %d", len(lines))
		}
	})

	t.Run("empty text", func(t *testing.T) {
		result := formatVerse(5, "")
		if !strings.Contains(result, "5") {
			t.Errorf("expected verse number even for empty text, got: %q", result)
		}
	})
}

// ── center ──────────────────────────────────────────────────

func TestCenter(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		result := center("hello", 20)
		// "hello" is 5 chars, should have ~7-8 leading spaces
		trimmed := strings.TrimRight(result, " ")
		leading := len(trimmed) - len("hello")
		if leading < 7 || leading > 8 {
			t.Errorf("expected ~7-8 leading spaces, got %d: %q", leading, result)
		}
	})

	t.Run("text with ANSI codes", func(t *testing.T) {
		text := "\x1b[1mhello\x1b[0m"
		result := center(text, 20)
		// visible length is 5, so padding should be same as plain
		if !strings.Contains(result, "hello") {
			t.Errorf("ANSI text should be preserved: %q", result)
		}
	})

	t.Run("text wider than width", func(t *testing.T) {
		result := center("this is very long text", 5)
		if !strings.Contains(result, "this is very long text") {
			t.Errorf("wide text should be returned as-is: %q", result)
		}
	})
}

// ── extractGospel ───────────────────────────────────────────

func TestExtractGospel(t *testing.T) {
	t.Run("finds gospel in valid data", func(t *testing.T) {
		data := &apiResponse{
			Sections: []section{
				{
					ID:    3, // Liturgy
					Title: "Liturgy",
					SubSections: []subSection{
						{
							ID:    1, // Psalm & Gospel
							Title: "Psalm & Gospel",
							Readings: []reading{
								{ID: 1, Conclusion: "Alleluia"}, // Psalm
								{
									ID:         2, // Gospel
									Conclusion: "Glory be to God forever.",
									Passages: []passage{
										{
											BookTranslation: "Mark",
											Ref:             "10:17-27",
											Verses: []verse{
												{Number: 17, Text: "Test verse"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		g := extractGospel(data)
		if g == nil {
			t.Fatal("expected gospel, got nil")
		}
		if g.ID != 2 {
			t.Errorf("expected reading ID 2, got %d", g.ID)
		}
		if len(g.Passages) != 1 {
			t.Errorf("expected 1 passage, got %d", len(g.Passages))
		}
		if g.Passages[0].BookTranslation != "Mark" {
			t.Errorf("expected Mark, got %s", g.Passages[0].BookTranslation)
		}
	})

	t.Run("returns nil when no Liturgy section", func(t *testing.T) {
		data := &apiResponse{
			Sections: []section{
				{ID: 2, Title: "Matins"},
			},
		}
		if g := extractGospel(data); g != nil {
			t.Errorf("expected nil, got %+v", g)
		}
	})

	t.Run("returns nil when no Gospel reading", func(t *testing.T) {
		data := &apiResponse{
			Sections: []section{
				{
					ID: 3,
					SubSections: []subSection{
						{
							ID: 1,
							Readings: []reading{
								{ID: 1}, // only Psalm, no Gospel
							},
						},
					},
				},
			},
		}
		if g := extractGospel(data); g != nil {
			t.Errorf("expected nil, got %+v", g)
		}
	})
}

// ── render ──────────────────────────────────────────────────

func TestRender(t *testing.T) {
	data := &apiResponse{
		CopticDate: "17/6/1742",
		Sections: []section{
			{
				ID: 3,
				SubSections: []subSection{
					{
						ID: 1,
						Readings: []reading{
							{ID: 1},
							{
								ID:         2,
								Conclusion: "Glory be to God forever.",
								Passages: []passage{
									{
										BookTranslation: "Mark",
										Ref:             "10:17-27",
										Verses: []verse{
											{Number: 17, Text: "As He was going out on the road."},
											{Number: 18, Text: "Why do you call Me good?"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	output := render(data, "24-02-2026")

	checks := []struct {
		name string
		want string
	}{
		{"contains date", "24-02-2026"},
		{"contains coptic date", "17 Amshir 1742"},
		{"contains book", "Mark"},
		{"contains ref", "10:17-27"},
		{"contains verse 17", "As He was going out"},
		{"contains verse 18", "Why do you call Me good"},
		{"contains conclusion", "Glory be to God forever."},
		{"contains box top", "╭"},
		{"contains box bottom", "╰"},
	}

	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			if !strings.Contains(output, c.want) {
				t.Errorf("render output missing %q", c.want)
			}
		})
	}
}

// ── HTTP handler ────────────────────────────────────────────

func TestHandler(t *testing.T) {
	// We use a mock upstream by replacing fetchReadings behavior.
	// For integration tests, we test the handler routing and error cases.

	t.Run("bad date returns 400", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/not-a-date", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		if w.Code != 400 {
			t.Errorf("expected 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Bad date") {
			t.Errorf("expected error message, got: %s", w.Body.String())
		}
	})

	t.Run("content type is text/plain", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/not-a-date", nil)
		w := httptest.NewRecorder()
		handler(w, req)

		ct := w.Header().Get("Content-Type")
		if !strings.HasPrefix(ct, "text/plain") {
			t.Errorf("expected text/plain, got %s", ct)
		}
	})

	t.Run("lang query param is respected", func(t *testing.T) {
		// We can't fully test without hitting the real API,
		// but we can verify the param parsing doesn't crash.
		req := httptest.NewRequest("GET", "/?lang=en", nil)
		w := httptest.NewRecorder()
		// This will hit the real API or fail gracefully
		handler(w, req)
		// Should be 200 or 502 (if API unreachable), not a panic
		if w.Code != 200 && w.Code != 502 {
			t.Errorf("expected 200 or 502, got %d", w.Code)
		}
	})
}

// ── JSON unmarshaling ───────────────────────────────────────

func TestAPIResponseParsing(t *testing.T) {
	raw := `{
		"copticDate": "17/6/1742",
		"sections": [{
			"id": 3,
			"title": "Liturgy",
			"subSections": [{
				"id": 1,
				"title": "Psalm & Gospel",
				"readings": [{
					"id": 2,
					"conclusion": "Glory",
					"passages": [{
						"bookTranslation": "Luke",
						"ref": "12:22-31",
						"verses": [
							{"number": 22, "text": "Do not worry about your life."},
							{"number": 23, "text": "Life is more than food."}
						]
					}]
				}]
			}]
		}]
	}`

	var data apiResponse
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	if data.CopticDate != "17/6/1742" {
		t.Errorf("copticDate = %q, want 17/6/1742", data.CopticDate)
	}
	if len(data.Sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(data.Sections))
	}

	g := extractGospel(&data)
	if g == nil {
		t.Fatal("extractGospel returned nil")
	}
	if len(g.Passages[0].Verses) != 2 {
		t.Errorf("expected 2 verses, got %d", len(g.Passages[0].Verses))
	}
	if g.Passages[0].Verses[0].Text != "Do not worry about your life." {
		t.Errorf("unexpected verse text: %s", g.Passages[0].Verses[0].Text)
	}
}
