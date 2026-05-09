#!/usr/bin/env bash
# Activates the project's git hooks by pointing core.hooksPath at .husky/.
# Idempotent — safe to re-run.
#
# Hooks live in .husky/ at the repo root (mixed Go + TS repo, single
# source of truth). Bypass any single commit with `git commit --no-verify`.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

if [[ ! -d .husky ]]; then
    echo "❌ install-hooks: .husky/ not found at repo root." >&2
    exit 1
fi

# Make the hook executable on the local checkout (file mode bit may be
# lost on Windows clones; chmod here is a no-op on a fresh Unix clone).
chmod +x .husky/pre-commit 2>/dev/null || true

git config core.hooksPath .husky
echo "✓ git core.hooksPath set to .husky/"
echo "  Pre-commit will run: AmE/BrE grep, golangci-lint, prettier, eslint."
echo "  Bypass for WIP: git commit --no-verify"
