#!/usr/bin/env bash
set -euo pipefail

if [[ $# -ne 1 ]]; then
  echo "usage: $0 ABSOLUTE_TEMPLATE_DIRECTORY" >&2
  exit 2
fi

destination=$1
case "$destination" in
  /*) ;;
  *) echo "template directory must be absolute" >&2; exit 2 ;;
esac
case "$destination" in
  /|/etc|/usr|/var|/home) echo "refusing broad template directory" >&2; exit 2 ;;
esac

repository_root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
revision=$(tr -d '[:space:]' < "$repository_root/config/nuclei-templates.lock")
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

archive="$work/upstream.tar.gz"
prefix="nuclei-templates-$revision"
curl --fail --silent --show-error --location \
  "https://codeload.github.com/projectdiscovery/nuclei-templates/tar.gz/$revision" \
  --output "$archive"
tar -xzf "$archive" -C "$work" \
  "$prefix/http/cves" \
  "$prefix/http/exposures" \
  "$prefix/http/misconfiguration" \
  "$prefix/http/technologies" \
  "$prefix/http/vulnerabilities" \
  "$prefix/ssl"

mkdir -p "$work/release"
cp -a "$work/$prefix/http" "$work/release/http"
cp -a "$work/$prefix/ssl" "$work/release/ssl"
printf '%s\n' "$revision" > "$work/release/CYBEREDGE_UPSTREAM_REVISION"

mkdir -p "$destination"
rsync -a --delete "$work/release/" "$destination/"

template_count=$(find "$destination" -type f \( -name '*.yaml' -o -name '*.yml' \) | wc -l)
echo "installed $template_count pinned templates at revision $revision"
