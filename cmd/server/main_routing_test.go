package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestExtracurricularRoutesMountedUnderV1 guards a B3 frontend slice
// contract: every extracurricular endpoint is reachable under
// /api/v1/extracurricular/... per RegisterExtracurricularRoutes
// docstring. The bug shipped in v0.166.0 mounted these routes under
// the bare /api group, returning 404 on every frontend call.
//
// We inspect main.go's source to confirm the call site wraps a
// /v1-prefixed group before handing it to RegisterExtracurricularRoutes.
// Static-source check avoids spinning up the full server, all DI
// dependencies of which are inappropriate for a routing regression.
func TestExtracurricularRoutesMountedUnderV1(t *testing.T) {
	src := readMainGo(t)

	callRx := regexp.MustCompile(`RegisterExtracurricularRoutes\(\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*,`)
	callMatches := callRx.FindAllStringSubmatch(src, -1)
	if len(callMatches) == 0 {
		t.Fatal("RegisterExtracurricularRoutes call not found in main.go")
	}

	groupVar := callMatches[0][1]
	if !groupVarResolvesToV1(src, groupVar) {
		t.Errorf("RegisterExtracurricularRoutes called with %q which does not resolve to a /v1-prefixed Gin group; "+
			"frontend expects /api/v1/extracurricular/events", groupVar)
	}
}

func readMainGo(t *testing.T) string {
	t.Helper()
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	body, err := os.ReadFile(filepath.Join(cwd, "main.go"))
	if err != nil {
		t.Fatalf("read main.go: %v", err)
	}
	return string(body)
}

// groupVarResolvesToV1 walks the chain of `name := parent.Group("...")`
// declarations starting from groupVar and returns true if any link in
// the chain carries a literal "/v1" path segment. Handles both direct
// declaration (apiV1 := router.Group("/api/v1")) and nested wrapping
// (v1 := protectedGroup.Group("/v1")).
func groupVarResolvesToV1(src, name string) bool {
	for range 5 {
		declRx := regexp.MustCompile(`(?m)^\s*` + regexp.QuoteMeta(name) + `\s*:?=\s*([a-zA-Z_][a-zA-Z0-9_]*)\.Group\("([^"]+)"\)`)
		m := declRx.FindStringSubmatch(src)
		if len(m) == 0 {
			return false
		}
		if strings.Contains(m[2], "/v1") {
			return true
		}
		name = m[1]
	}
	return false
}
