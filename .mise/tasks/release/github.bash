#!/usr/bin/env bash
#MISE description="build release artifacts"
set -o pipefail -o errexit -o nounset

git stash
git switch main
git pull --all --tags
last_version=$(git describe --tags --always --abbrev=7 2>/dev/null || git rev-parse --short=7 HEAD)
echo "Last version: $last_version"
read -r -p "Enter new version: " new_version
echo "New version: $new_version"
mise run release:artifacts "$new_version"
release_files=()
while IFS= read -r -d '' f; do release_files+=("$f"); done < <(find "./dist/$new_version" -type f -print0)
gh release create "$new_version" "${release_files[@]}" --generate-notes
