package fuzzy

// Span is a contiguous haystack slice. Adjacent runes with the same Matched
// flag are merged. Concatenating every Text in order equals the haystack.
type Span struct {
	Text    string `json:"text"`
	Matched bool   `json:"matched"`
}

// Result is the outcome of Match or MatchAll.
type Result struct {
	OK    bool
	Score int
	Spans []Span
}
