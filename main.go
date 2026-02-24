package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ── ANSI ────────────────────────────────────────────────────

const (
	bold   = "\x1b[1m"
	dim    = "\x1b[2m"
	italic = "\x1b[3m"
	rst    = "\x1b[0m"
	gold   = "\x1b[38;5;178m"
	gray   = "\x1b[38;5;240m"
	white  = "\x1b[38;5;255m"
)

// ── Layout ──────────────────────────────────────────────────

const (
	W      = 60
	numW   = 3
	gap    = 2
	prefix = 2 + numW + gap // "  17  " = 7 visible chars
	textW  = W - prefix
)

// ── Config ──────────────────────────────────────────────────

var langMap = map[string]int{
	"fr": 1, "en": 2, "ar": 3, "it": 4,
	"de": 6, "pl": 7, "es": 8, "nl": 9,
}

var copticMonths = []string{
	"Thout", "Paopi", "Hathor", "Koiak", "Tobi", "Amshir",
	"Baramhat", "Baramouda", "Bashans", "Paoni", "Epep",
	"Mesori", "Nasie",
}

// ── API types ───────────────────────────────────────────────

type apiResponse struct {
	Sections  []section `json:"sections"`
	CopticDate string   `json:"copticDate"`
}

type section struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	SubSections []subSection `json:"subSections"`
}

type subSection struct {
	ID       int       `json:"id"`
	Title    string    `json:"title"`
	Readings []reading `json:"readings"`
}

type reading struct {
	ID         int       `json:"id"`
	Conclusion string    `json:"conclusion"`
	Passages   []passage `json:"passages"`
}

type passage struct {
	BookTranslation string  `json:"bookTranslation"`
	Ref             string  `json:"ref"`
	Verses          []verse `json:"verses"`
}

type verse struct {
	Number int    `json:"number"`
	Text   string `json:"text"`
}

// ── Cache ───────────────────────────────────────────────────

type cacheEntry struct {
	data      *apiResponse
	fetchedAt time.Time
}

var (
	cache   = map[string]cacheEntry{}
	cacheMu sync.Mutex
)

func fetchReadings(date string, langID int) (*apiResponse, error) {
	key := fmt.Sprintf("%s:%d", date, langID)

	cacheMu.Lock()
	if e, ok := cache[key]; ok && time.Since(e.fetchedAt) < 24*time.Hour {
		cacheMu.Unlock()
		return e.data, nil
	}
	cacheMu.Unlock()

	url := fmt.Sprintf(
		"https://api.katameros.app/readings/gregorian/%s?languageId=%d",
		date, langID,
	)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data apiResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	cacheMu.Lock()
	cache[key] = cacheEntry{data: &data, fetchedAt: time.Now()}
	cacheMu.Unlock()

	return &data, nil
}

// ── Language detection ──────────────────────────────────────

func detectLang(header string) int {
	if header == "" {
		return 1
	}

	type langQ struct {
		lang string
		q    float64
	}

	var parts []langQ
	for _, chunk := range strings.Split(header, ",") {
		chunk = strings.TrimSpace(chunk)
		q := 1.0
		tag := chunk
		if i := strings.Index(chunk, ";q="); i >= 0 {
			tag = chunk[:i]
			if v, err := strconv.ParseFloat(chunk[i+3:], 64); err == nil {
				q = v
			}
		}
		base := strings.ToLower(strings.SplitN(strings.TrimSpace(tag), "-", 2)[0])
		parts = append(parts, langQ{lang: base, q: q})
	}

	sort.Slice(parts, func(i, j int) bool { return parts[i].q > parts[j].q })

	for _, p := range parts {
		if id, ok := langMap[p.lang]; ok {
			return id
		}
	}
	return 1
}

// ── Formatting ──────────────────────────────────────────────

func formatCoptic(s string) string {
	parts := strings.SplitN(s, "/", 3)
	if len(parts) != 3 {
		return s
	}
	d, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	if m < 1 || m > 13 {
		return s
	}
	return fmt.Sprintf("%d %s %s", d, copticMonths[m-1], parts[2])
}

func wordWrap(text string, width int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var lines []string
	line := words[0]

	for _, w := range words[1:] {
		if len(line)+1+len(w) > width {
			lines = append(lines, line)
			line = w
		} else {
			line += " " + w
		}
	}
	lines = append(lines, line)
	return lines
}

func formatVerse(n int, text string) string {
	text = strings.TrimSpace(text)
	num := fmt.Sprintf("%s%*d%s", gray, numW, n, rst)
	indent := strings.Repeat(" ", prefix)
	lines := wordWrap(text, textW)
	if len(lines) == 0 {
		return fmt.Sprintf("  %s", num)
	}

	var b strings.Builder
	fmt.Fprintf(&b, "  %s%s%s", num, strings.Repeat(" ", gap), lines[0])
	for _, l := range lines[1:] {
		fmt.Fprintf(&b, "\n%s%s", indent, l)
	}
	return b.String()
}

func center(text string, width int) string {
	// strip ANSI to measure visible length
	visible := text
	for {
		i := strings.Index(visible, "\x1b[")
		if i < 0 {
			break
		}
		j := strings.IndexByte(visible[i:], 'm')
		if j < 0 {
			break
		}
		visible = visible[:i] + visible[i+j+1:]
	}

	pad := width - len(visible)
	if pad <= 0 {
		return text
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// ── Extract & render ────────────────────────────────────────

func extractGospel(data *apiResponse) *reading {
	for _, sec := range data.Sections {
		if sec.ID != 3 { // Liturgy
			continue
		}
		for _, sub := range sec.SubSections {
			if sub.ID != 1 { // Psalm & Gospel
				continue
			}
			for i := range sub.Readings {
				if sub.Readings[i].ID == 2 { // Gospel
					return &sub.Readings[i]
				}
			}
		}
	}
	return nil
}

func render(data *apiResponse, date string) string {
	var b strings.Builder
	coptic := formatCoptic(data.CopticDate)
	inner := W - 6 // content width inside box padding
	dash := inner + 2

	// ── Header box ──

	dateLine := fmt.Sprintf("%s%s%s  ·  %s%s", gold, bold, date, coptic, rst)

	fmt.Fprintln(&b)
	fmt.Fprintf(&b, "  %s╭%s╮%s\n", gray, strings.Repeat("─", dash), rst)
	fmt.Fprintf(&b, "  %s│%s %s %s│%s\n", gray, rst, center(dateLine, inner), gray, rst)
	fmt.Fprintf(&b, "  %s╰%s╯%s\n", gray, strings.Repeat("─", dash), rst)
	fmt.Fprintln(&b)

	// ── Gospel ──

	g := extractGospel(data)
	if g == nil || len(g.Passages) == 0 {
		fmt.Fprintf(&b, "  %sNo Gospel reading found for this date.%s\n\n", dim, rst)
		return b.String()
	}

	for _, p := range g.Passages {
		title := fmt.Sprintf("%s %s", p.BookTranslation, p.Ref)
		fmt.Fprintf(&b, "  %s%s%s%s\n", white, bold, title, rst)
		fmt.Fprintf(&b, "  %s%s%s\n", gray, strings.Repeat("─", len(title)), rst)
		fmt.Fprintln(&b)

		for _, v := range p.Verses {
			fmt.Fprintln(&b, formatVerse(v.Number, v.Text))
		}
		fmt.Fprintln(&b)
	}

	// ── Conclusion ──

	if g.Conclusion != "" {
		rule := fmt.Sprintf("%s───%s", gray, rst)
		fmt.Fprintln(&b, center(rule, W))
		fmt.Fprintln(&b)
		conclusion := fmt.Sprintf("%s%s%s%s", dim, italic, g.Conclusion, rst)
		fmt.Fprintln(&b, center(conclusion, W))
		fmt.Fprintln(&b)
	}

	return b.String()
}

// ── HTTP handler ────────────────────────────────────────────

func handler(w http.ResponseWriter, r *http.Request) {
	// Language
	langID := 1
	if q := r.URL.Query().Get("lang"); q != "" {
		if id, ok := langMap[strings.ToLower(q)]; ok {
			langID = id
		}
	} else {
		langID = detectLang(r.Header.Get("Accept-Language"))
	}

	// Date
	path := strings.TrimPrefix(r.URL.Path, "/")
	date := path
	if date == "" {
		date = time.Now().Format("02-01-2006")
	}

	// Validate date
	if _, err := time.Parse("02-01-2006", date); err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		fmt.Fprintf(w, "\n  %sBad date: %s (use dd-mm-yyyy)%s\n\n", dim, date, rst)
		return
	}

	// Fetch
	data, err := fetchReadings(date, langID)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(502)
		fmt.Fprintf(w, "\n  %sKatameros API unreachable. Try again later.%s\n\n", dim, rst)
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, render(data, date))
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}

	http.HandleFunc("/", handler)
	fmt.Fprintf(os.Stderr, "katameros-cli listening on :%s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
