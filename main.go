package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

// ── ANSI ────────────────────────────────────────────────────

var (
	aBold   = "\x1b[1m"
	aDim    = "\x1b[2m"
	aItalic = "\x1b[3m"
	aRst    = "\x1b[0m"
	aGold   = "\x1b[38;5;178m"
	aGray   = "\x1b[38;5;240m"
	aWhite  = "\x1b[38;5;255m"
	aCyan   = "\x1b[38;5;109m"
)

func disableColors() {
	aBold = ""
	aDim = ""
	aItalic = ""
	aRst = ""
	aGold = ""
	aGray = ""
	aWhite = ""
	aCyan = ""
}

// ── Layout ──────────────────────────────────────────────────

const (
	W      = 60
	numW   = 3
	gap    = 2
	prefix = 2 + numW + gap
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
	Sections   []section `json:"sections"`
	CopticDate string    `json:"copticDate"`
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
	ID           int       `json:"id"`
	Title        string    `json:"title"`
	Introduction string    `json:"introduction"`
	Conclusion   string    `json:"conclusion"`
	Passages     []passage `json:"passages"`
	HTML         string    `json:"html"`
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

// ── Fetch ───────────────────────────────────────────────────

func fetchReadings(date string, langID int) (*apiResponse, error) {
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
	return &data, nil
}

// ── View filter ─────────────────────────────────────────────

// A filter identifies a piece of the API response to extract.
// -1 means "match all" for that level.
type filter struct {
	sectionID    int
	subSectionID int
	readingID    int
}

var viewFilters = map[string][]filter{
	"gospel":     {{3, 1, 2}},
	"psalm":      {{3, 1, 1}},
	"synaxarium": {{3, 10, -1}},
	"synax":      {{3, 10, -1}},
	"pauline":    {{3, 3, -1}},
	"catholic":   {{3, 4, -1}},
	"epistles":   {{3, 3, -1}, {3, 4, -1}},
	"acts":       {{3, 5, -1}},
	"prophecies": {{2, 2, -1}},
	"matins":     {{2, -1, -1}},
	"liturgy":    {{3, -1, -1}},
	"all":        {{-1, -1, -1}},
}

func resolveViews(names []string) []filter {
	var filters []filter
	seen := map[filter]bool{}
	for _, name := range names {
		if ff, ok := viewFilters[strings.ToLower(name)]; ok {
			for _, f := range ff {
				if !seen[f] {
					seen[f] = true
					filters = append(filters, f)
				}
			}
		}
	}
	return filters
}

func matches(f filter, secID, subID, readID int) bool {
	if f.sectionID != -1 && f.sectionID != secID {
		return false
	}
	if f.subSectionID != -1 && f.subSectionID != subID {
		return false
	}
	if f.readingID != -1 && f.readingID != readID {
		return false
	}
	return true
}

func anyMatch(filters []filter, secID, subID, readID int) bool {
	for _, f := range filters {
		if matches(f, secID, subID, readID) {
			return true
		}
	}
	return false
}

// ── Formatting helpers ──────────────────────────────────────

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
	num := fmt.Sprintf("%s%*d%s", aGray, numW, n, aRst)
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

var ansiRe = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func visibleLen(s string) int {
	return len(ansiRe.ReplaceAllString(s, ""))
}

func center(text string, width int) string {
	vis := visibleLen(text)
	pad := width - vis
	if pad <= 0 {
		return text
	}
	left := pad / 2
	right := pad - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

// ── HTML stripping ──────────────────────────────────────────

var (
	htmlTagRe    = regexp.MustCompile(`<[^>]*>`)
	htmlEntityRe = regexp.MustCompile(`&[a-zA-Z]+;`)
	multiSpaceRe = regexp.MustCompile(`[^\S\n]{2,}`)
	multiNewline = regexp.MustCompile(`\n{3,}`)
)

var entityMap = map[string]string{
	"&amp;": "&", "&lt;": "<", "&gt;": ">",
	"&quot;": "\"", "&apos;": "'", "&nbsp;": " ",
}

func stripHTML(s string) string {
	// Replace <p> and <br> with newlines
	s = strings.ReplaceAll(s, "</p>", "\n\n")
	s = strings.ReplaceAll(s, "<br>", "\n")
	s = strings.ReplaceAll(s, "<br/>", "\n")
	s = strings.ReplaceAll(s, "<br />", "\n")
	// Strip all remaining tags
	s = htmlTagRe.ReplaceAllString(s, "")
	// Decode common entities
	for ent, repl := range entityMap {
		s = strings.ReplaceAll(s, ent, repl)
	}
	s = htmlEntityRe.ReplaceAllString(s, "")
	// Clean up whitespace
	s = multiSpaceRe.ReplaceAllString(s, " ")
	s = multiNewline.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

// ── Render ──────────────────────────────────────────────────

func renderHeader(b *strings.Builder, date string, copticDate string) {
	coptic := formatCoptic(copticDate)
	inner := W - 6
	dash := inner + 2
	dateLine := fmt.Sprintf("%s%s%s  ·  %s%s", aGold, aBold, date, coptic, aRst)

	fmt.Fprintln(b)
	fmt.Fprintf(b, "  %s╭%s╮%s\n", aGray, strings.Repeat("─", dash), aRst)
	fmt.Fprintf(b, "  %s│%s %s %s│%s\n", aGray, aRst, center(dateLine, inner), aGray, aRst)
	fmt.Fprintf(b, "  %s╰%s╯%s\n", aGray, strings.Repeat("─", dash), aRst)
	fmt.Fprintln(b)
}

func renderSectionHeader(b *strings.Builder, title string) {
	line := fmt.Sprintf("── %s ──", title)
	fmt.Fprintf(b, "  %s%s%s%s\n\n", aCyan, aBold, line, aRst)
}

func renderSubSectionHeader(b *strings.Builder, title string) {
	fmt.Fprintf(b, "  %s▸ %s%s\n\n", aDim, title, aRst)
}

func renderPassageReading(b *strings.Builder, rd reading) {
	for _, p := range rd.Passages {
		title := fmt.Sprintf("%s %s", p.BookTranslation, p.Ref)
		fmt.Fprintf(b, "  %s%s%s%s\n", aWhite, aBold, title, aRst)
		fmt.Fprintf(b, "  %s%s%s\n\n", aGray, strings.Repeat("─", len(title)), aRst)
		for _, v := range p.Verses {
			fmt.Fprintln(b, formatVerse(v.Number, v.Text))
		}
		fmt.Fprintln(b)
	}
	if rd.Conclusion != "" {
		rule := fmt.Sprintf("%s───%s", aGray, aRst)
		fmt.Fprintln(b, center(rule, W))
		fmt.Fprintln(b)
		// Wrap long conclusions
		lines := wordWrap(rd.Conclusion, W-4)
		for _, l := range lines {
			styled := fmt.Sprintf("%s%s%s%s", aDim, aItalic, l, aRst)
			fmt.Fprintln(b, center(styled, W))
		}
		fmt.Fprintln(b)
	}
}

func renderSynaxariumReading(b *strings.Builder, rd reading) {
	if rd.Title != "" {
		fmt.Fprintf(b, "  %s%s%s%s\n", aWhite, aBold, rd.Title, aRst)
		fmt.Fprintf(b, "  %s%s%s\n\n", aGray, strings.Repeat("─", len(rd.Title)), aRst)
	}
	if rd.HTML != "" {
		plain := stripHTML(rd.HTML)
		for _, para := range strings.Split(plain, "\n\n") {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}
			lines := wordWrap(para, W-4)
			for _, l := range lines {
				fmt.Fprintf(b, "  %s\n", l)
			}
			fmt.Fprintln(b)
		}
	}
}

func render(data *apiResponse, date string, filters []filter) string {
	var b strings.Builder
	renderHeader(&b, date, data.CopticDate)

	multiSection := false
	for _, f := range filters {
		if f.sectionID == -1 || (len(filters) > 1) {
			multiSection = true
			break
		}
	}

	found := false
	for _, sec := range data.Sections {
		// Check if any reading in this section matches
		sectionHasMatch := false
		for _, sub := range sec.SubSections {
			for _, rd := range sub.Readings {
				if anyMatch(filters, sec.ID, sub.ID, rd.ID) {
					sectionHasMatch = true
					break
				}
			}
			if sectionHasMatch {
				break
			}
		}
		if !sectionHasMatch {
			continue
		}

		if multiSection {
			renderSectionHeader(&b, sec.Title)
		}

		for _, sub := range sec.SubSections {
			subHasMatch := false
			for _, rd := range sub.Readings {
				if anyMatch(filters, sec.ID, sub.ID, rd.ID) {
					subHasMatch = true
					break
				}
			}
			if !subHasMatch {
				continue
			}

			if multiSection {
				renderSubSectionHeader(&b, sub.Title)
			}

			for _, rd := range sub.Readings {
				if !anyMatch(filters, sec.ID, sub.ID, rd.ID) {
					continue
				}
				found = true
				if rd.HTML != "" {
					renderSynaxariumReading(&b, rd)
				} else if len(rd.Passages) > 0 {
					renderPassageReading(&b, rd)
				}
			}
		}
	}

	if !found {
		fmt.Fprintf(&b, "  %sNo readings found for this date.%s\n\n", aDim, aRst)
	}

	return b.String()
}

// ── CLI ─────────────────────────────────────────────────────

const helpText = `katameros-cli — daily Coptic Orthodox readings in your terminal

Usage:
  katameros-cli [options] [section...]

Sections (combinable, default: gospel):
  gospel        Liturgy Gospel
  psalm         Liturgy Psalm
  synaxarium    Saint of the day (alias: synax)
  pauline       Pauline Epistle
  catholic      Catholic Epistle
  epistles      Pauline + Catholic combined
  acts          Acts of the Apostles
  prophecies    Matins OT readings
  matins        Full Matins section
  liturgy       Full Liturgy section
  all           Everything

Options:
  -d, --date    Date in dd-mm-yyyy format (default: today)
  -l, --lang    Language code (default: fr)
  --no-color    Disable colored output
  -h, --help    Show this help

Languages:
  fr  French       en  English      ar  Arabic
  it  Italian      de  German       pl  Polish
  es  Spanish      nl  Dutch

Examples:
  katameros-cli
  katameros-cli -l en
  katameros-cli -d 25-12-2025 gospel synaxarium
  katameros-cli all -l ar
`

type cliArgs struct {
	date     string
	lang     string
	noColor  bool
	help     bool
	sections []string
}

func parseCLI(args []string) cliArgs {
	c := cliArgs{
		date: time.Now().Format("02-01-2006"),
		lang: "fr",
	}

	i := 0
	for i < len(args) {
		a := args[i]
		switch a {
		case "-h", "--help":
			c.help = true
		case "--no-color":
			c.noColor = true
		case "-d", "--date":
			if i+1 < len(args) {
				i++
				c.date = args[i]
			}
		case "-l", "--lang":
			if i+1 < len(args) {
				i++
				c.lang = strings.ToLower(args[i])
			}
		default:
			if strings.HasPrefix(a, "-") {
				fmt.Fprintf(os.Stderr, "Unknown option: %s\n", a)
				fmt.Fprintf(os.Stderr, "Run 'katameros-cli --help' for usage.\n")
				os.Exit(1)
			}
			c.sections = append(c.sections, a)
		}
		i++
	}

	if len(c.sections) == 0 {
		c.sections = []string{"gospel"}
	}

	return c
}

func main() {
	c := parseCLI(os.Args[1:])

	if c.help {
		fmt.Print(helpText)
		return
	}

	// Color detection
	if c.noColor || os.Getenv("NO_COLOR") != "" {
		disableColors()
	} else if !term.IsTerminal(int(os.Stdout.Fd())) {
		disableColors()
	}

	// Validate date
	if _, err := time.Parse("02-01-2006", c.date); err != nil {
		fmt.Fprintf(os.Stderr, "Bad date: %s (use dd-mm-yyyy)\n", c.date)
		os.Exit(1)
	}

	// Resolve language
	langID, ok := langMap[c.lang]
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown language: %s\n", c.lang)
		fmt.Fprintf(os.Stderr, "Available: fr, en, ar, it, de, pl, es, nl\n")
		os.Exit(1)
	}

	// Resolve views
	filters := resolveViews(c.sections)
	if len(filters) == 0 {
		fmt.Fprintf(os.Stderr, "Unknown section: %s\n", strings.Join(c.sections, ", "))
		fmt.Fprintf(os.Stderr, "Run 'katameros-cli --help' for available sections.\n")
		os.Exit(1)
	}

	// Fetch
	data, err := fetchReadings(c.date, langID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching readings: %v\n", err)
		os.Exit(1)
	}

	// Render
	fmt.Print(render(data, c.date, filters))
}
