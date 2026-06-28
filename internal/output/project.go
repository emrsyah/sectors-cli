package output

import (
	"encoding/json"
	"strings"
)

// These helpers shape a response client-side, before it reaches an agent's
// context window: Project keeps only requested fields, Truncate caps list
// length, and Count reduces a response to just its size. Each operates on the
// raw JSON body and returns new JSON; on any parse error the original body is
// returned unchanged (the CLI never corrupts a payload it can't shape).

type seg struct {
	key   string
	array bool // segment is an array to map over, e.g. `results[]`
}

func splitPath(p string) []seg {
	parts := strings.Split(p, ".")
	segs := make([]seg, 0, len(parts))
	for _, part := range parts {
		s := seg{key: part}
		if strings.HasSuffix(part, "[]") {
			s.key = strings.TrimSuffix(part, "[]")
			s.array = true
		}
		segs = append(segs, s)
	}
	return segs
}

// Project returns a new JSON document containing only the requested dotted
// paths. Use `a.b` to descend objects and `key[]` to map over an array
// (e.g. `results[].symbol`, or `[].symbol` for a top-level array). Paths that
// don't exist are skipped.
func Project(body []byte, paths []string) ([]byte, error) {
	doc, ok := decodeForPaths(body, paths)
	if !ok {
		// Not JSON we can shape — return the body untouched (the CLI never
		// corrupts a payload it can't parse).
		return body, nil
	}
	var out any
	for _, p := range paths {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if v, ok := projectPath(doc, splitPath(p)); ok {
			out = deepMerge(out, v)
		}
	}
	if out == nil {
		out = map[string]any{}
	}
	return json.Marshal(out)
}

// decodeForPaths materializes just enough of body to satisfy paths. For a
// top-level object it shallow-decodes the keys (each value stays a raw slice)
// and then fully parses only the branches the paths reference — projectPath
// never reads sibling keys, so pruning them up front is transparent but avoids
// allocating the entire document. Other shapes fall back to a full decode.
func decodeForPaths(body []byte, paths []string) (any, bool) {
	if firstJSONByte(body) != '{' {
		var doc any
		if err := json.Unmarshal(body, &doc); err != nil {
			return nil, false
		}
		return doc, true
	}

	var top map[string]json.RawMessage
	if err := json.Unmarshal(body, &top); err != nil {
		return nil, false
	}
	doc := make(map[string]any, len(paths))
	for _, p := range paths {
		key := splitPath(strings.TrimSpace(p))[0].key
		if key == "" {
			continue // top-level array segment can't match an object
		}
		if _, done := doc[key]; done {
			continue
		}
		raw, ok := top[key]
		if !ok {
			continue // referenced key absent → projectPath would skip it anyway
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			continue
		}
		doc[key] = v
	}
	return doc, true
}

// firstJSONByte returns the first non-whitespace byte of b, or 0 if none.
func firstJSONByte(b []byte) byte {
	for _, c := range b {
		switch c {
		case ' ', '\t', '\n', '\r':
		default:
			return c
		}
	}
	return 0
}

// projectPath rebuilds the subtree along segs, preserving structure.
func projectPath(node any, segs []seg) (any, bool) {
	if len(segs) == 0 {
		return node, true
	}
	s := segs[0]
	rest := segs[1:]

	// Top-level array segment: `[]` with no key.
	if s.key == "" && s.array {
		arr, ok := node.([]any)
		if !ok {
			return nil, false
		}
		return mapArray(arr, rest), true
	}

	m, ok := node.(map[string]any)
	if !ok {
		return nil, false
	}
	child, ok := m[s.key]
	if !ok {
		return nil, false
	}

	if s.array {
		arr, ok := child.([]any)
		if !ok {
			return nil, false
		}
		return map[string]any{s.key: mapArray(arr, rest)}, true
	}
	v, ok := projectPath(child, rest)
	if !ok {
		return nil, false
	}
	return map[string]any{s.key: v}, true
}

func mapArray(arr []any, rest []seg) []any {
	out := make([]any, 0, len(arr))
	for _, el := range arr {
		if v, ok := projectPath(el, rest); ok {
			out = append(out, v)
		}
	}
	return out
}

// deepMerge combines two projected fragments so multiple --select paths produce
// one coherent tree (objects merge by key, arrays merge element-wise).
func deepMerge(dst, src any) any {
	if dst == nil {
		return src
	}
	dm, ok1 := dst.(map[string]any)
	sm, ok2 := src.(map[string]any)
	if ok1 && ok2 {
		for k, v := range sm {
			if ex, ok := dm[k]; ok {
				dm[k] = deepMerge(ex, v)
			} else {
				dm[k] = v
			}
		}
		return dm
	}
	da, ok1 := dst.([]any)
	sa, ok2 := src.([]any)
	if ok1 && ok2 {
		n := len(da)
		if len(sa) > n {
			n = len(sa)
		}
		out := make([]any, n)
		for i := 0; i < n; i++ {
			switch {
			case i < len(da) && i < len(sa):
				out[i] = deepMerge(da[i], sa[i])
			case i < len(da):
				out[i] = da[i]
			default:
				out[i] = sa[i]
			}
		}
		return out
	}
	return src
}

// Truncate caps the primary list to max items. For object-wrapped lists
// (`{results:[…]}` / `{data:[…]}`) it adds `"_truncated": true` when it cut
// anything; bare top-level arrays are simply shortened.
func Truncate(body []byte, max int) ([]byte, error) {
	if max < 0 {
		return body, nil
	}
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return body, nil
	}
	switch v := doc.(type) {
	case []any:
		if len(v) > max {
			return json.Marshal(v[:max])
		}
	case map[string]any:
		for _, k := range []string{"results", "data"} {
			if arr, ok := v[k].([]any); ok {
				if len(arr) > max {
					v[k] = arr[:max]
					v["_truncated"] = true
					return json.Marshal(v)
				}
				break
			}
		}
	}
	return body, nil
}

// Count reduces a response to {"count": n}, where n is the length of the
// primary list (or 1 for a single object).
func Count(body []byte) ([]byte, error) {
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		return body, nil
	}
	n := 1
	switch v := doc.(type) {
	case []any:
		n = len(v)
	case map[string]any:
		for _, k := range []string{"results", "data"} {
			if arr, ok := v[k].([]any); ok {
				n = len(arr)
				break
			}
		}
	}
	return json.Marshal(map[string]any{"count": n})
}
