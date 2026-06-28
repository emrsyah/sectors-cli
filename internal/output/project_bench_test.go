package output

import (
	"encoding/json"
	"testing"
)

// reportBody builds a report-shaped object (~50KB) with several large
// top-level branches, mirroring `idx company report`. --select typically asks
// for one or two of these branches.
func reportBody(tb testing.TB) []byte {
	tb.Helper()
	branch := func(n int) map[string]any {
		m := make(map[string]any, n)
		for i := 0; i < n; i++ {
			m[string(rune('a'+i%26))+itoa(i)] = map[string]any{
				"value": 12345.678, "label": "some descriptive label text", "i": i,
			}
		}
		return m
	}
	doc := map[string]any{
		"symbol":       "BBCA.JK",
		"company_name": "PT Bank Central Asia Tbk",
		"overview":     branch(40),
		"valuation":    branch(120),
		"future":       branch(120),
		"financials":   branch(200),
		"dividend":     branch(120),
	}
	b, err := json.Marshal(doc)
	if err != nil {
		tb.Fatal(err)
	}
	return b
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var buf [20]byte
	p := len(buf)
	for i > 0 {
		p--
		buf[p] = byte('0' + i%10)
		i /= 10
	}
	return string(buf[p:])
}

// BenchmarkProjectSelect projects a single top-level branch out of a large
// report object — the common `--select overview` case. The lazy decode should
// avoid materializing the un-selected branches.
func BenchmarkProjectSelect(b *testing.B) {
	body := reportBody(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := Project(body, []string{"overview"}); err != nil {
			b.Fatal(err)
		}
	}
}

// TestProject_PrunesUnreferencedBranches confirms the lazy decode returns the
// same result as a full decode: only selected branches survive, others vanish.
func TestProject_PrunesUnreferencedBranches(t *testing.T) {
	in := []byte(`{"a":{"x":1},"b":{"y":2},"c":[1,2,3]}`)
	out, err := Project(in, []string{"a", "c"})
	if err != nil {
		t.Fatal(err)
	}
	got := mustJSON(t, out).(map[string]any)
	if _, ok := got["a"]; !ok {
		t.Error("selected key a missing")
	}
	if _, ok := got["c"]; !ok {
		t.Error("selected key c missing")
	}
	if _, ok := got["b"]; ok {
		t.Error("unselected key b should be pruned")
	}
	if got["a"].(map[string]any)["x"].(float64) != 1 {
		t.Errorf("a.x wrong: %v", got)
	}
}

// TestProject_InvalidJSONUnchanged keeps the contract that an unparseable body
// is returned verbatim.
func TestProject_InvalidJSONUnchanged(t *testing.T) {
	in := []byte(`{not valid json`)
	out, err := Project(in, []string{"a"})
	if err != nil {
		t.Fatal(err)
	}
	if string(out) != string(in) {
		t.Errorf("invalid JSON should pass through unchanged: got %q", out)
	}
}
