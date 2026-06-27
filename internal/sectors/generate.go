// Package sectors contains the generated, typed client for the Sectors
// Financial API v2 plus thin hand-written helpers (auth, base URL).
//
// The *.gen.go files are produced by oapi-codegen from ../../sectors-schema.json.
// Regenerate with `go generate ./...` after updating the spec.
package sectors

//go:generate go run ../../tools/fixspec -in ../../sectors-schema.json -out ../../sectors-schema.fixed.json
//go:generate go tool oapi-codegen -config ../../oapi-codegen.yaml ../../sectors-schema.fixed.json
