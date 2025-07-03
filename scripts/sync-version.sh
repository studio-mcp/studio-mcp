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

echo "Version sync complete!"

# Note: This script only updates package.json in-place for GoReleaser
# It does not commit changes - that's handled by the release workflow
