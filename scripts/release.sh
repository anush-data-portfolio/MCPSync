#!/usr/bin/env bash
# Usage: scripts/release.sh v0.2.0
# Bumps version in all tracked files, commits, tags, and pushes.
# The push triggers .github/workflows/release.yml which builds binaries,
# creates the GitHub Release, and publishes to npm.
set -euo pipefail

VERSION="${1:-}"
if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version>   (e.g. $0 v0.2.0)"
  exit 1
fi

# Strip leading 'v' for files that store bare semver (package.json, CITATION.cff)
BARE="${VERSION#v}"

# Validate semver shape: vMAJOR.MINOR.PATCH (pre-release suffixes allowed)
if ! [[ "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9] ]]; then
  echo "Version must start with vMAJOR.MINOR.PATCH (got: $VERSION)"
  exit 1
fi

# Must be on main with a clean working tree
BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" != "main" ]]; then
  echo "Must be on main (currently on: $BRANCH)"
  exit 1
fi
if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "Working tree is dirty — commit or stash changes first."
  exit 1
fi

# Tag must not already exist
if git rev-parse "$VERSION" >/dev/null 2>&1; then
  echo "Tag $VERSION already exists."
  exit 1
fi

echo "Releasing $VERSION..."

TODAY="$(date +%Y-%m-%d)"

# ── 1. package.json ────────────────────────────────────────────────────────────
npm version "$BARE" --no-git-tag-version --allow-same-version --silent
echo "  ✓ package.json → $BARE"

# ── 2. CITATION.cff ───────────────────────────────────────────────────────────
# Replace the bare semver on the 'version:' line
sed -i "s/^version: .*/version: $BARE/" CITATION.cff
# Update release date
sed -i "s/^date-released: .*/date-released: \"$TODAY\"/" CITATION.cff
echo "  ✓ CITATION.cff → $BARE ($TODAY)"

# ── 3. CHANGELOG.md ───────────────────────────────────────────────────────────
# Inject a new "Unreleased" section header placeholder after the first ---
# so the author fills in the notes before pushing (or it stays as a stub).
CHANGELOG_ENTRY="## [$BARE] — $TODAY"
if grep -q "^\[$BARE\]" CHANGELOG.md 2>/dev/null; then
  echo "  ℹ CHANGELOG.md already has an entry for $BARE — skipping injection"
else
  # Insert after the first "---" separator line
  sed -i "0,/^---$/{s/^---$/---\n\n${CHANGELOG_ENTRY}\n\n### Changed\n\n- *(add release notes here)*\n\n---/}" CHANGELOG.md
  echo "  ✓ CHANGELOG.md — added stub for $BARE (fill in release notes)"
fi

# ── 4. Commit version bump ─────────────────────────────────────────────────────
git add package.json CITATION.cff CHANGELOG.md
git commit -m "chore: release $VERSION"
echo "  ✓ committed version bump"

# ── 5. Tag and push ────────────────────────────────────────────────────────────
git tag "$VERSION"
git push origin main
git push origin "$VERSION"
echo ""
echo "  Pushed tag $VERSION → release workflow will build binaries, create GitHub Release, and publish to npm."
echo "  Track progress: https://github.com/anush-data-portfolio/MCPSync/actions"
