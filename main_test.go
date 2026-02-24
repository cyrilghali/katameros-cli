package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ── parseCLI ────────────────────────────────────────────────

func TestParseCLI(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantLang string
		wantSec  []string
		wantHelp bool
	}{
		{"no args defaults to gospel+fr", nil, "fr", []string{"gospel"}, false},
		{"help flag", []string{"--help"}, "fr", []string{"gospel"}, true},
		{"short help", []string{"-h"}, "fr", []string{"gospel"}, true},
		{"language flag", []string{"-l", "en"}, "en", []string{"gospel"}, false},
		{"long language flag", []string{"--lang", "ar"}, "ar", []string{"gospel"}, false},
		{"date flag", []string{"-d", "25-12-2025"}, "fr", []string{"gospel"}, false},
		{"single section", []string{"synaxarium"}, "fr", []string{"synaxarium"}, false},
		{"multiple sections", []string{"gospel", "synaxarium"}, "fr", []string{"gospel", "synaxarium"}, false},
		{"sections with flags", []string{"-l", "en", "epistles", "acts"}, "en", []string{"epistles", "acts"}, false},
		{"flags after sections", []string{"all", "-l", "ar"}, "ar", []string{"all"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := parseCLI(tt.args)
			if c.lang != tt.wantLang {
				t.Errorf("lang = %q, want %q", c.lang, tt.wantLang)
			}
			if c.help != tt.wantHelp {
				t.Errorf("help = %v, want %v", c.help, tt.wantHelp)
			}
			if len(c.sections) != len(tt.wantSec) {
				t.Fatalf("sections = %v, want %v", c.sections, tt.wantSec)
			}
			for i, s := range c.sections {
				if s != tt.wantSec[i] {
					t.Errorf("sections[%d] = %q, want %q", i, s, tt.wantSec[i])
				}
			}
		})
	}
}

// ── resolveViews ────────────────────────────────────────────

func TestResolveViews(t *testing.T) {
	tests := []struct {
		name  string
		names []string
		want  int // expected number of filters
	}{
		{"gospel", []string{"gospel"}, 1},
		{"epistles expands to 2", []string{"epistles"}, 2},
		{"synax alias", []string{"synax"}, 1},
		{"combined no dupes", []string{"gospel", "gospel"}, 1},
		{"unknown returns empty", []string{"bogus"}, 0},
		{"mixed known+unknown", []string{"gospel", "bogus"}, 1},
		{"all", []string{"all"}, 1},
		{"multiple views", []string{"gospel", "synaxarium", "acts"}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveViews(tt.names)
			if len(got) != tt.want {
				t.Errorf("resolveViews(%v) returned %d filters, want %d", tt.names, len(got), tt.want)
			}
		})
	}
}

// ── matches / anyMatch ──────────────────────────────────────

func TestMatches(t *testing.T) {
	tests := []struct {
		name    string
		f       filter
		sec     int
		sub     int
		read    int
		want    bool
	}{
		{"exact match", filter{3, 1, 2}, 3, 1, 2, true},
		{"wrong section", filter{3, 1, 2}, 2, 1, 2, false},
		{"wildcard section", filter{-1, 1, 2}, 3, 1, 2, true},
		{"wildcard sub", filter{3, -1, 2}, 3, 5, 2, true},
		{"wildcard reading", filter{3, 1, -1}, 3, 1, 99, true},
		{"all wildcards", filter{-1, -1, -1}, 1, 2, 3, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matches(tt.f, tt.sec, tt.sub, tt.read)
			if got != tt.want {
				t.Errorf("matches(%v, %d, %d, %d) = %v, want %v",
					tt.f, tt.sec, tt.sub, tt.read, got, tt.want)
			}
		})
	}
}

// ── formatCoptic ────────────────────────────────────────────

func TestFormatCoptic(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"17/6/1742", "17 Amshir 1742"},
		{"1/1/1740", "1 Thout 1740"},
		{"30/12/1742", "30 Mesori 1742"},
		{"5/13/1742", "5 Nasie 1742"},
		{"bad", "bad"},
		{"1/14/1742", "1/14/1742"},
		{"1/0/1742", "1/0/1742"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
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
		want  int
	}{
		{"empty", "", 40, 0},
		{"short fits", "Hello world", 40, 1},
		{"wraps", "Hello world this is a longer sentence that should wrap", 20, 3},
		{"single long word", "superlongword", 5, 1},
		{"whitespace only", "   ", 40, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wordWrap(tt.text, tt.width)
			if len(got) != tt.want {
				t.Errorf("wordWrap(%q, %d) = %d lines, want %d: %v",
					tt.text, tt.width, len(got), tt.want, got)
			}
		})
	}
}

// ── formatVerse ─────────────────────────────────────────────

func TestFormatVerse(t *testing.T) {
	t.Run("short verse", func(t *testing.T) {
		r := formatVerse(1, "In the beginning.")
		if !strings.Contains(r, "In the beginning.") {
			t.Errorf("missing text: %q", r)
		}
	})

	t.Run("long verse wraps", func(t *testing.T) {
		r := formatVerse(17, strings.Repeat("word ", 20))
		if strings.Count(r, "\n") < 1 {
			t.Error("expected multiple lines for long verse")
		}
	})

	t.Run("empty text", func(t *testing.T) {
		r := formatVerse(5, "")
		if !strings.Contains(r, "5") {
			t.Errorf("expected verse number: %q", r)
		}
	})
}

// ── center ──────────────────────────────────────────────────

func TestCenter(t *testing.T) {
	t.Run("plain text", func(t *testing.T) {
		r := center("hello", 20)
		leading := len(r) - len(strings.TrimLeft(r, " "))
		if leading < 7 || leading > 8 {
			t.Errorf("expected ~7-8 leading spaces, got %d", leading)
		}
	})

	t.Run("with ANSI", func(t *testing.T) {
		r := center("\x1b[1mhi\x1b[0m", 20)
		if !strings.Contains(r, "hi") {
			t.Errorf("lost text: %q", r)
		}
	})

	t.Run("wider than width", func(t *testing.T) {
		r := center("very long text here", 5)
		if !strings.Contains(r, "very long text here") {
			t.Errorf("should return as-is: %q", r)
		}
	})
}

// ── stripHTML ───────────────────────────────────────────────

func TestStripHTML(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"plain text", "hello", "hello"},
		{"p tags", "<p>Hello</p><p>World</p>", "Hello\n\nWorld"},
		{"br tags", "line1<br>line2", "line1\nline2"},
		{"nested tags", "<p><strong>Bold</strong> and <em>italic</em></p>", "Bold and italic"},
		{"entities", "&amp; &lt; &gt;", "& < >"},
		{"span with attributes", `<span dir="RTL" lang="AR-SA">text</span>`, "text"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stripHTML(tt.input)
			if got != tt.want {
				t.Errorf("stripHTML(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ── extractGospel (via filters) ─────────────────────────────

func buildTestData() *apiResponse {
	return &apiResponse{
		CopticDate: "17/6/1742",
		Sections: []section{
			{
				ID:    2,
				Title: "Matins",
				SubSections: []subSection{
					{
						ID:    2,
						Title: "Prophecies",
						Readings: []reading{
							{ID: 7, Passages: []passage{{BookTranslation: "Job", Ref: "19:1-26", Verses: []verse{{Number: 1, Text: "Job spoke."}}}}},
						},
					},
					{
						ID:    1,
						Title: "Psalm & Gospel",
						Readings: []reading{
							{ID: 1, Passages: []passage{{BookTranslation: "Psalms", Ref: "41:4", Verses: []verse{{Number: 4, Text: "Psalm verse."}}}}},
							{ID: 2, Conclusion: "Glory.", Passages: []passage{{BookTranslation: "Luke", Ref: "12:22-31", Verses: []verse{{Number: 22, Text: "Do not worry."}}}}},
						},
					},
				},
			},
			{
				ID:    3,
				Title: "Liturgy",
				SubSections: []subSection{
					{ID: 3, Title: "Pauline Epistle", Readings: []reading{{ID: 3, Passages: []passage{{BookTranslation: "2 Corinthians", Ref: "9:6-15", Verses: []verse{{Number: 6, Text: "Sow bountifully."}}}}}}},
					{ID: 4, Title: "Catholic Epistle", Readings: []reading{{ID: 4, Passages: []passage{{BookTranslation: "James", Ref: "1:1-12", Verses: []verse{{Number: 1, Text: "James, servant."}}}}}}},
					{ID: 5, Title: "Acts", Readings: []reading{{ID: 5, Passages: []passage{{BookTranslation: "Acts", Ref: "4:13-22", Verses: []verse{{Number: 13, Text: "Boldness of Peter."}}}}}}},
					{ID: 10, Title: "Synaxarium", Readings: []reading{{ID: 6, Title: "St. Mina", HTML: "<p>He was a monk.</p>"}}},
					{
						ID:    1,
						Title: "Psalm & Gospel",
						Readings: []reading{
							{ID: 1, Passages: []passage{{BookTranslation: "Psalms", Ref: "41:1", Verses: []verse{{Number: 1, Text: "Blessed."}}}}},
							{ID: 2, Conclusion: "Glory be.", Passages: []passage{{BookTranslation: "Mark", Ref: "10:17-27", Verses: []verse{{Number: 17, Text: "Good Teacher."}}}}},
						},
					},
				},
			},
		},
	}
}

func TestExtractGospel(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"gospel"})
	out := render(data, "24-02-2026", filters)

	if !strings.Contains(out, "Mark") {
		t.Error("gospel should contain Mark")
	}
	if strings.Contains(out, "Job") {
		t.Error("gospel should not contain Job")
	}
	if strings.Contains(out, "Matins") {
		t.Error("single view should not show section headers")
	}
}

func TestExtractSynaxarium(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"synaxarium"})
	out := render(data, "24-02-2026", filters)

	if !strings.Contains(out, "St. Mina") {
		t.Error("synaxarium should contain title")
	}
	if !strings.Contains(out, "He was a monk.") {
		t.Error("synaxarium should contain stripped HTML content")
	}
}

func TestExtractEpistles(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"epistles"})
	out := render(data, "24-02-2026", filters)

	if !strings.Contains(out, "2 Corinthians") {
		t.Error("epistles should contain Pauline")
	}
	if !strings.Contains(out, "James") {
		t.Error("epistles should contain Catholic")
	}
}

func TestExtractAll(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"all"})
	out := render(data, "24-02-2026", filters)

	for _, want := range []string{"Job", "Luke", "2 Corinthians", "James", "Acts", "St. Mina", "Mark"} {
		if !strings.Contains(out, want) {
			t.Errorf("all should contain %q", want)
		}
	}
	if !strings.Contains(out, "Matins") {
		t.Error("all should show section headers")
	}
	if !strings.Contains(out, "Liturgy") {
		t.Error("all should show section headers")
	}
}

func TestCombinedViews(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"gospel", "synaxarium"})
	out := render(data, "24-02-2026", filters)

	if !strings.Contains(out, "Mark") {
		t.Error("combined should contain gospel")
	}
	if !strings.Contains(out, "St. Mina") {
		t.Error("combined should contain synaxarium")
	}
	// Multi-view shows section headers
	if !strings.Contains(out, "Liturgy") {
		t.Error("combined views should show section headers")
	}
}

func TestRenderHeader(t *testing.T) {
	data := buildTestData()
	filters := resolveViews([]string{"gospel"})
	out := render(data, "24-02-2026", filters)

	if !strings.Contains(out, "24-02-2026") {
		t.Error("missing date")
	}
	if !strings.Contains(out, "17 Amshir 1742") {
		t.Error("missing coptic date")
	}
	if !strings.Contains(out, "╭") || !strings.Contains(out, "╰") {
		t.Error("missing box borders")
	}
}

// ── JSON parsing ────────────────────────────────────────────

func TestAPIResponseParsing(t *testing.T) {
	raw := `{
		"copticDate": "17/6/1742",
		"sections": [{
			"id": 3, "title": "Liturgy",
			"subSections": [{
				"id": 1, "title": "Psalm & Gospel",
				"readings": [{
					"id": 2, "conclusion": "Glory",
					"passages": [{"bookTranslation": "Luke", "ref": "12:22-31",
						"verses": [{"number": 22, "text": "Do not worry."}]
					}]
				}]
			}]
		}]
	}`

	var data apiResponse
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if data.CopticDate != "17/6/1742" {
		t.Errorf("copticDate = %q", data.CopticDate)
	}

	filters := resolveViews([]string{"gospel"})
	out := render(&data, "24-02-2026", filters)
	if !strings.Contains(out, "Do not worry.") {
		t.Error("render missing verse text")
	}
}
