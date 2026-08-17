// Package markdown provides lossless, in-memory editing of Markdown sections.
package markdown

import (
	"errors"
	"strings"
)

var (
	ErrInvalidHeader    = errors.New("invalid markdown section header")
	ErrSectionNotFound  = errors.New("markdown section not found")
	ErrSectionExists    = errors.New("markdown section already exists")
	ErrAmbiguousSection = errors.New("markdown section is ambiguous")
)

// Document is a Markdown document whose untouched bytes are preserved exactly.
type Document struct {
	source string
}

// Parse creates an editable Markdown document.
func Parse(source string) (*Document, error) {
	return &Document{source: source}, nil
}

// String returns the current document text.
func (d *Document) String() string {
	return d.source
}

// GetSectionContent returns the body of the uniquely matching section.
func (d *Document) GetSectionContent(header string) (content string, found bool, err error) {
	selector, err := parseSelector(header)
	if err != nil {
		return "", false, err
	}
	section, found, err := d.findSection(selector)
	if err != nil || !found {
		return "", found, err
	}
	return d.source[section.contentStart:section.end], true, nil
}

// ReplaceSectionContent replaces the body of the uniquely matching section.
func (d *Document) ReplaceSectionContent(header, content string) error {
	selector, err := parseSelector(header)
	if err != nil {
		return err
	}
	section, found, err := d.findSection(selector)
	if err != nil {
		return err
	}
	if !found {
		return ErrSectionNotFound
	}

	lineEnding := section.lineEnding
	if lineEnding == "" {
		lineEnding = firstLineEnding(d.source)
	}
	replacement := normalizeContent(content, lineEnding)
	prefix := d.source[:section.contentStart]
	if section.lineEnding == "" {
		prefix += lineEnding
	}
	d.source = prefix + replacement + d.source[section.end:]
	return nil
}

// RemoveSection removes the heading and body of the uniquely matching section.
func (d *Document) RemoveSection(header string) error {
	selector, err := parseSelector(header)
	if err != nil {
		return err
	}
	section, found, err := d.findSection(selector)
	if err != nil {
		return err
	}
	if !found {
		return ErrSectionNotFound
	}
	d.source = d.source[:section.start] + d.source[section.end:]
	return nil
}

// RemoveAllSections removes every section matching header. Unlike
// RemoveSection, duplicate matches are expected and removed together. A
// missing header is a successful no-op.
func (d *Document) RemoveAllSections(header string) (int, error) {
	selector, err := parseSelector(header)
	if err != nil {
		return 0, err
	}
	headings := scanHeadings(d.source)
	spans := make([]section, 0)
	for i, candidate := range headings {
		if candidate.level != selector.level || candidate.title != selector.title {
			continue
		}
		end := len(d.source)
		for _, boundary := range headings[i+1:] {
			if boundary.level <= candidate.level {
				end = boundary.start
				break
			}
		}
		spans = append(spans, section{heading: candidate, end: end})
	}
	for i := len(spans) - 1; i >= 0; i-- {
		span := spans[i]
		d.source = d.source[:span.start] + d.source[span.end:]
	}
	return len(spans), nil
}

// InsertBeforeFirstSection inserts a new section before the first recognized
// heading, or appends it when the document contains no headings.
func (d *Document) InsertBeforeFirstSection(header, content string) error {
	selector, err := parseSelector(header)
	if err != nil {
		return err
	}
	_, found, err := d.findSection(selector)
	if err != nil {
		return err
	}
	if found {
		return ErrSectionExists
	}

	headings := scanHeadings(d.source)
	insertAt := len(d.source)
	if len(headings) > 0 {
		insertAt = headings[0].start
	}
	lineEnding := firstLineEnding(d.source)
	inserted := header + lineEnding + normalizeContent(content, lineEnding)
	separator := ""
	if insertAt == len(d.source) && insertAt > 0 && !endsWithLineEnding(d.source) {
		separator = lineEnding
	}
	d.source = d.source[:insertAt] + separator + inserted + d.source[insertAt:]
	return nil
}

type heading struct {
	level        int
	title        string
	start        int
	contentStart int
	lineEnding   string
}

type section struct {
	heading
	end int
}

func (d *Document) findSection(selector heading) (section, bool, error) {
	headings := scanHeadings(d.source)
	matches := make([]int, 0, 1)
	for i, candidate := range headings {
		if candidate.level == selector.level && candidate.title == selector.title {
			matches = append(matches, i)
		}
	}
	if len(matches) == 0 {
		return section{}, false, nil
	}
	if len(matches) > 1 {
		return section{}, false, ErrAmbiguousSection
	}

	i := matches[0]
	end := len(d.source)
	for _, candidate := range headings[i+1:] {
		if candidate.level <= headings[i].level {
			end = candidate.start
			break
		}
	}
	return section{heading: headings[i], end: end}, true, nil
}

func parseSelector(value string) (heading, error) {
	if strings.ContainsAny(value, "\r\n") || len(value) < 3 || value[0] != '#' {
		return heading{}, ErrInvalidHeader
	}
	i := 0
	for i < len(value) && value[i] == '#' {
		i++
	}
	if i < 1 || i > 6 || i >= len(value) || value[i] != ' ' {
		return heading{}, ErrInvalidHeader
	}
	title := normalizeHeadingTitle(value[i+1:])
	if title == "" {
		return heading{}, ErrInvalidHeader
	}
	return heading{level: i, title: title}, nil
}

func scanHeadings(source string) []heading {
	var headings []heading
	var fence fenceState
	for start := 0; start < len(source); {
		textEnd, next, ending := splitLine(source, start)
		line := source[start:textEnd]
		if fence.active {
			if isFenceClose(line, fence) {
				fence = fenceState{}
			}
		} else if opened, ok := parseFenceOpen(line); ok {
			fence = opened
		} else if level, title, ok := parseATXHeading(line); ok {
			headings = append(headings, heading{
				level: level, title: title, start: start,
				contentStart: next, lineEnding: ending,
			})
		}
		start = next
	}
	return headings
}

func parseATXHeading(line string) (int, string, bool) {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	startHashes := i
	for i < len(line) && line[i] == '#' {
		i++
	}
	level := i - startHashes
	if level < 1 || level > 6 {
		return 0, "", false
	}
	if i < len(line) && line[i] != ' ' && line[i] != '\t' {
		return 0, "", false
	}
	title := ""
	if i < len(line) {
		title = normalizeHeadingTitle(line[i+1:])
	}
	return level, title, true
}

func normalizeHeadingTitle(value string) string {
	title := strings.Trim(value, " \t")
	if title == "" {
		return ""
	}
	end := len(title)
	for end > 0 && title[end-1] == '#' {
		end--
	}
	if end < len(title) && end > 0 && (title[end-1] == ' ' || title[end-1] == '\t') {
		title = strings.TrimRight(title[:end], " \t")
	}
	return title
}

type fenceState struct {
	active bool
	char   byte
	count  int
}

func parseFenceOpen(line string) (fenceState, bool) {
	i := leadingFenceIndent(line)
	if i < 0 || i >= len(line) || (line[i] != '`' && line[i] != '~') {
		return fenceState{}, false
	}
	ch := line[i]
	j := i
	for j < len(line) && line[j] == ch {
		j++
	}
	if j-i < 3 {
		return fenceState{}, false
	}
	if ch == '`' && strings.ContainsRune(line[j:], '`') {
		return fenceState{}, false
	}
	return fenceState{active: true, char: ch, count: j - i}, true
}

func isFenceClose(line string, fence fenceState) bool {
	i := leadingFenceIndent(line)
	if i < 0 || i >= len(line) || line[i] != fence.char {
		return false
	}
	j := i
	for j < len(line) && line[j] == fence.char {
		j++
	}
	return j-i >= fence.count && strings.Trim(line[j:], " \t") == ""
}

func leadingFenceIndent(line string) int {
	i := 0
	for i < len(line) && i < 3 && line[i] == ' ' {
		i++
	}
	if i < len(line) && line[i] == ' ' {
		return -1
	}
	return i
}

func splitLine(source string, start int) (textEnd, next int, ending string) {
	for i := start; i < len(source); i++ {
		if source[i] == '\n' {
			if i > start && source[i-1] == '\r' {
				return i - 1, i + 1, "\r\n"
			}
			return i, i + 1, "\n"
		}
	}
	return len(source), len(source), ""
}

func firstLineEnding(source string) string {
	for i := 0; i < len(source); i++ {
		if source[i] == '\n' {
			if i > 0 && source[i-1] == '\r' {
				return "\r\n"
			}
			return "\n"
		}
	}
	return "\n"
}

func normalizeContent(content, lineEnding string) string {
	if content == "" || endsWithLineEnding(content) {
		return content
	}
	return content + lineEnding
}

func endsWithLineEnding(value string) bool {
	return strings.HasSuffix(value, "\n")
}
