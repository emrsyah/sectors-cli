package output

import (
	"encoding/json"
	"testing"
)

func mustJSON(t *testing.T, b []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		t.Fatalf("invalid JSON %q: %v", b, err)
	}
	return v
}

func TestProject_NestedObjectPaths(t *testing.T) {
	in := []byte(`{"overview":{"market_cap":100,"sector":"X"},"valuation":{"pe":9},"junk":[1,2]}`)
	out, err := Project(in, []string{"overview.market_cap", "valuation"})
	if err != nil {
		t.Fatal(err)
	}
	got := mustJSON(t, out).(map[string]any)
	ov := got["overview"].(map[string]any)
	if ov["market_cap"].(float64) != 100 {
		t.Errorf("market_cap missing: %v", got)
	}
	if _, ok := ov["sector"]; ok {
		t.Error("sector should have been projected away")
	}
	if _, ok := got["valuation"]; !ok {
		t.Error("valuation subtree missing")
	}
	if _, ok := got["junk"]; ok {
		t.Error("junk should not be present")
	}
}

func TestProject_ArrayWildcardMergesFields(t *testing.T) {
	in := []byte(`{"results":[{"symbol":"A","close":1,"x":9},{"symbol":"B","close":2,"x":9}],"total":2}`)
	out, err := Project(in, []string{"results[].symbol", "results[].close"})
	if err != nil {
		t.Fatal(err)
	}
	got := mustJSON(t, out).(map[string]any)
	rows := got["results"].([]any)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	r0 := rows[0].(map[string]any)
	if r0["symbol"] != "A" || r0["close"].(float64) != 1 {
		t.Errorf("row0 = %v", r0)
	}
	if _, ok := r0["x"]; ok {
		t.Error("x should be projected away")
	}
	if _, ok := got["total"]; ok {
		t.Error("total not selected, should be absent")
	}
}

func TestProject_TopLevelArrayWildcard(t *testing.T) {
	in := []byte(`[{"symbol":"A","v":1},{"symbol":"B","v":2}]`)
	out, err := Project(in, []string{"[].symbol"})
	if err != nil {
		t.Fatal(err)
	}
	rows := mustJSON(t, out).([]any)
	if len(rows) != 2 || rows[0].(map[string]any)["symbol"] != "A" {
		t.Errorf("got %v", rows)
	}
}

func TestProject_MissingPathSkipped(t *testing.T) {
	in := []byte(`{"a":1}`)
	out, err := Project(in, []string{"nope.missing", "a"})
	if err != nil {
		t.Fatal(err)
	}
	got := mustJSON(t, out).(map[string]any)
	if got["a"].(float64) != 1 || len(got) != 1 {
		t.Errorf("got %v", got)
	}
}

func TestTruncate_BareArray(t *testing.T) {
	out, err := Truncate([]byte(`[1,2,3,4,5]`), 2)
	if err != nil {
		t.Fatal(err)
	}
	arr := mustJSON(t, out).([]any)
	if len(arr) != 2 {
		t.Errorf("want 2, got %d", len(arr))
	}
}

func TestTruncate_ResultsWrapperAddsMarker(t *testing.T) {
	out, err := Truncate([]byte(`{"results":[1,2,3],"has_next":true}`), 1)
	if err != nil {
		t.Fatal(err)
	}
	got := mustJSON(t, out).(map[string]any)
	if len(got["results"].([]any)) != 1 {
		t.Errorf("results not truncated: %v", got)
	}
	if got["_truncated"] != true {
		t.Errorf("_truncated marker missing: %v", got)
	}
}

func TestTruncate_NoOpWhenUnderLimit(t *testing.T) {
	in := []byte(`{"results":[1,2]}`)
	out, _ := Truncate(in, 5)
	got := mustJSON(t, out).(map[string]any)
	if _, ok := got["_truncated"]; ok {
		t.Error("should not mark when nothing truncated")
	}
}

func TestCount(t *testing.T) {
	cases := map[string]int{
		`[1,2,3]`:            3,
		`{"results":[1,2]}`:  2,
		`{"data":[1,2,3,4]}`: 4,
		`{"symbol":"BBCA"}`:  1,
	}
	for in, want := range cases {
		out, err := Count([]byte(in))
		if err != nil {
			t.Fatal(err)
		}
		got := mustJSON(t, out).(map[string]any)
		if int(got["count"].(float64)) != want {
			t.Errorf("Count(%s) = %v, want %d", in, got["count"], want)
		}
	}
}
