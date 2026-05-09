#!/usr/bin/env bash
# Smoke-test the pre-commit hook against each violation class it
# should catch. Creates throwaway files with a known violation,
# stages them, runs the hook directly (not `git commit` — we don't
# want to actually commit), then resets the index. Asserts the hook
# exits non-zero with a recognisable error string for each case.
#
# Run: bash _tools/test-pre-commit.sh
# Exits 0 if every case caught, non-zero otherwise.

set -uo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -x .husky/pre-commit ]]; then
    echo "❌ .husky/pre-commit not found or not executable" >&2
    exit 1
fi

PASS=0
FAIL=0

cleanup() {
    git reset --quiet HEAD -- "$@" 2>/dev/null || true
    rm -f "$@" 2>/dev/null || true
}

run_case() {
    local name="$1"
    local expect_substr="$2"
    local file="$3"
    shift 3
    local content="$*"

    printf '%s' "$content" > "$file"
    git add "$file" 2>/dev/null

    local out
    out=$(.husky/pre-commit 2>&1)
    local rc=$?

    cleanup "$file"

    if [[ $rc -eq 0 ]]; then
        echo "❌ FAIL: '$name' — hook exited 0 but should have rejected" >&2
        FAIL=$((FAIL + 1))
        return
    fi
    if ! grep -qF "$expect_substr" <<<"$out"; then
        echo "❌ FAIL: '$name' — exit non-zero but missing '$expect_substr' in output" >&2
        echo "    Got: $out" | head -5 >&2
        FAIL=$((FAIL + 1))
        return
    fi
    echo "✓ PASS: $name"
    PASS=$((PASS + 1))
}

# Violation 1 — BrE form in Go comment.
run_case "Go BrE 'behaviour'" \
    "BrE forms detected" \
    "_tmp_pre_commit_test_bre.go" \
    "package handlers" $'\n' "// behaviour smoke" $'\n'

# Violation 2 — BrE form 'defence' in Go comment.
run_case "Go BrE 'defence'" \
    "BrE forms detected" \
    "_tmp_pre_commit_test_def.go" \
    "package handlers" $'\n' "// defence-in-depth wrong spelling" $'\n'

# Violation 3 — prettier failure on a frontend test file.
# Prettier is configured with printWidth=100 and singleQuote=true; a
# long unbroken line with double quotes should fail --check.
LONG_LINE='const x = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa";'
run_case "Frontend prettier reject" \
    "prettier --check failed" \
    "frontend/src/_tmp_pre_commit_test.ts" \
    "$LONG_LINE"

echo ""
echo "Pre-commit smoke: $PASS passed, $FAIL failed"
exit $FAIL
