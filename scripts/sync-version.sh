#!/bin/bash

# Script to sync package.json version with git tags
# Usage: ./scripts/sync-version.sh [version]
# If no version is provided, it will use the latest git tag

set -e

# Get the directory of this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Function to get the latest git tag
get_latest_tag() {
    git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.1"
}

# Function to extract version from tag (remove 'v' prefix if present)
extract_version() {
    echo "$1" | sed 's/^v//'
}

# Get version from argument or latest tag
if [ -n "$1" ]; then
    VERSION="$1"
    # Add 'v' prefix if not present for git tag
    if [[ ! "$VERSION" =~ ^v ]]; then
        TAG="v$VERSION"
    else
        TAG="$VERSION"
        VERSION=$(extract_version "$VERSION")
    fi
else
    TAG=$(get_latest_tag)
    VERSION=$(extract_version "$TAG")
fi

echo "Syncing version to: $VERSION (tag: $TAG)"

# Update package.json version
cd "$PROJECT_ROOT"
if command -v jq >/dev/null 2>&1; then
    # Use jq if available
    jq --arg version "$VERSION" '.version = $version' package.json > package.json.tmp
    mv package.json.tmp package.json
else
    # Fallback to sed
    sed -i.bak "s/\"version\": \"[^\"]*\"/\"version\": \"$VERSION\"/" package.json
    rm -f package.json.bak
fi

echo "Updated package.json version to: $VERSION"

# If we're creating a new version, create/update the git tag
if [ -n "$1" ]; then
    echo "Creating git tag: $TAG"
    git tag -a "$TAG" -m "Release $TAG" || echo "Tag $TAG already exists"
fi

echo "Version sync complete!"
echo "To release:"
echo "  1. Commit any changes: git add package.json && git commit -m 'Bump version to $VERSION'"
echo "  2. Push tag: git push origin $TAG"
echo "  3. GoReleaser will handle the rest via GitHub Actions"
