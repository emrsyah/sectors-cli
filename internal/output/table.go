package output

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// maxCellWidth caps how wide a single rendered cell can be before truncation.
const maxCellWidth = 48

// renderTable turns a JSON list-of-records payload into an ASCII table. It
// returns (table, true) when the body is tabular, and ("", false) otherwise so
// the caller can fall back to JSON. Tabular shapes are:
//   - a top-level array of objects
//   - an object with a "results" (or "data") array of objects (the paginated
//     list shape), in which case remaining scalar keys are shown as a footer
func renderTable(body []byte) (string, bool) {
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return "", false
	}

	var rows []map[string]any
	var footer map[string]any

	switch v := doc.(type) {
	case []any:
		rows = toRows(v)
		if rows == nil {
			return "", false
		}
	case map[string]any:
		listKey := ""
		for _, k := range []string{"results", "data"} {
			if arr, ok := v[k].([]any); ok {
				if r := toRows(arr); r != nil {
					rows, listKey = r, k
					break
				}
			}
		}
		if listKey == "" {
			return "", false
		}
		// Surviving scalar keys (pagination metadata) become a footer.
		footer = map[string]any{}
		for k, val := range v {
			if k == listKey {
				continue
			}
			if isScalar(val) {
				footer[k] = val
			}
		}
	default:
		return "", false
	}

	if len(rows) == 0 {
		return "(no results)\n", true
	}

	cols := columns(rows)
	var b strings.Builder
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers(cols...)
	for _, row := range rows {
		cells := make([]string, len(cols))
		for i, c := range cols {
			cells[i] = cell(row[c])
		}
		t.Row(cells...)
	}
	b.WriteString(t.String())
	b.WriteString("\n")

	if len(footer) > 0 {
		keys := make([]string, 0, len(footer))
		for k := range footer {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = fmt.Sprintf("%s=%v", k, footer[k])
		}
		b.WriteString(strings.Join(parts, "  ") + "\n")
	}
	return b.String(), true
}

// toRows returns the slice as []map[string]any if every element is an object,
// else nil (not a uniform record list).
func toRows(arr []any) []map[string]any {
	rows := make([]map[string]any, 0, len(arr))
	for _, e := range arr {
		m, ok := e.(map[string]any)
		if !ok {
			return nil
		}
		rows = append(rows, m)
	}
	return rows
}

// columns returns the alphabetically-sorted union of keys across all rows.
func columns(rows []map[string]any) []string {
	set := map[string]struct{}{}
	for _, r := range rows {
		for k := range r {
			set[k] = struct{}{}
		}
	}
	cols := make([]string, 0, len(set))
	for k := range set {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}

func isScalar(v any) bool {
	switch v.(type) {
	case map[string]any, []any:
		return false
	default:
		return true
	}
}

// cell renders a single value: scalars verbatim, nested values as compact JSON,
// all truncated to maxCellWidth.
func cell(v any) string {
	var s string
	switch t := v.(type) {
	case nil:
		s = ""
	case string:
		s = t
	case float64:
		// Avoid scientific notation and trailing ".0" for integers.
		if t == float64(int64(t)) {
			s = fmt.Sprintf("%d", int64(t))
		} else {
			s = fmt.Sprintf("%g", t)
		}
	case bool:
		s = fmt.Sprintf("%t", t)
	default:
		raw, _ := json.Marshal(t)
		s = string(raw)
	}
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > maxCellWidth {
		s = s[:maxCellWidth-1] + "…"
	}
	return s
}
