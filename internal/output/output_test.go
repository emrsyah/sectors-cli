package output

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestParseFormat(t *testing.T) {
	for _, s := range []string{"auto", "json", "pretty", "table"} {
		if _, err := ParseFormat(s); err != nil {
			t.Errorf("ParseFormat(%q) unexpected error: %v", s, err)
		}
	}
	if _, err := ParseFormat("xml"); err == nil {
		t.Error("ParseFormat(\"xml\") = nil error, want error")
	}
}

func TestEmitJSON_Compact(t *testing.T) {
	var buf bytes.Buffer
	in := []byte(`{ "a": 1,  "b": [2, 3] }`)
	if err := EmitJSON(&buf, in, FormatJSON); err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(buf.String())
	if got != `{"a":1,"b":[2,3]}` {
		t.Errorf("compact = %q", got)
	}
}

func TestEmitJSON_Pretty(t *testing.T) {
	var buf bytes.Buffer
	if err := EmitJSON(&buf, []byte(`{"a":1}`), FormatPretty); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\n  \"a\": 1") {
		t.Errorf("pretty output not indented:\n%s", buf.String())
	}
}

func TestEmitJSON_TableArray(t *testing.T) {
	var buf bytes.Buffer
	in := []byte(`[{"symbol":"BBCA","close":6175},{"symbol":"BMRI","close":5000}]`)
	if err := EmitJSON(&buf, in, FormatTable); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{"symbol", "close", "BBCA", "6175", "BMRI"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
}

func TestEmitJSON_TableFallsBackForNestedObject(t *testing.T) {
	var buf bytes.Buffer
	// A single nested object is not tabular → should fall back to pretty JSON.
	in := []byte(`{"overview":{"sector":"Financials"}}`)
	if err := EmitJSON(&buf, in, FormatTable); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "\"overview\":") && !strings.Contains(buf.String(), "\"overview\"") {
		t.Errorf("expected JSON fallback, got:\n%s", buf.String())
	}
	if strings.Contains(buf.String(), "┌") {
		t.Errorf("nested object should not render as a table:\n%s", buf.String())
	}
}

func TestRenderTable_ResultsWrapperWithFooter(t *testing.T) {
	in := []byte(`{"results":[{"symbol":"D05"}],"has_next":true,"total":42}`)
	out, ok := renderTable(in)
	if !ok {
		t.Fatal("expected tabular render for results wrapper")
	}
	if !strings.Contains(out, "D05") {
		t.Errorf("missing row data:\n%s", out)
	}
	// Footer surfaces the pagination scalars.
	if !strings.Contains(out, "has_next=true") || !strings.Contains(out, "total=42") {
		t.Errorf("missing pagination footer:\n%s", out)
	}
}

func TestRenderTable_EmptyResults(t *testing.T) {
	out, ok := renderTable([]byte(`[]`))
	if !ok || !strings.Contains(out, "no results") {
		t.Errorf("empty array: ok=%v out=%q", ok, out)
	}
}

func TestRenderTable_NonTabular(t *testing.T) {
	if _, ok := renderTable([]byte(`"just a string"`)); ok {
		t.Error("scalar should not be tabular")
	}
	if _, ok := renderTable([]byte(`[1,2,3]`)); ok {
		t.Error("array of scalars should not be tabular")
	}
}

func TestCell(t *testing.T) {
	cases := map[string]struct {
		in   any
		want string
	}{
		"int float":   {float64(6175), "6175"},
		"real float":  {float64(3.14), "3.14"},
		"bool":        {true, "true"},
		"null":        {nil, ""},
		"nested json": {map[string]any{"a": float64(1)}, `{"a":1}`},
	}
	for name, c := range cases {
		if got := cell(c.in); got != c.want {
			t.Errorf("%s: cell(%v) = %q, want %q", name, c.in, got, c.want)
		}
	}
}

// A long multibyte (UTF-8) string must truncate on a rune boundary, never
// mid-rune — otherwise the table prints invalid UTF-8.
func TestCell_TruncatesMultibyteSafely(t *testing.T) {
	// 60 "é" runes (2 bytes each) exceeds the 48-rune cap.
	long := strings.Repeat("é", 60)
	got := cell(long)

	if !utf8.ValidString(got) {
		t.Fatalf("truncated cell is not valid UTF-8: %q", got)
	}
	if !strings.HasSuffix(got, "…") {
		t.Errorf("expected ellipsis suffix, got %q", got)
	}
	if n := utf8.RuneCountInString(got); n != maxCellWidth {
		t.Errorf("rune count = %d, want %d (%d content + ellipsis)", n, maxCellWidth, maxCellWidth-1)
	}
}
