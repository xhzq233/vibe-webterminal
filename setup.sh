#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
if [ "$(uname -s)" != Darwin ]; then echo 'This setup script is for macOS. See README for manual builds.'; exit 1; fi
if ! xcode-select -p >/dev/null 2>&1; then echo 'Run xcode-select --install first.'; exit 1; fi
missing=()
for package in go ttyd herdr; do
  if ! command -v "$package" >/dev/null 2>&1; then missing+=("$package"); fi
done
if [ "${#missing[@]}" -gt 0 ]; then
  if ! command -v brew >/dev/null; then echo 'Install Homebrew first: https://brew.sh'; exit 1; fi
  brew install "${missing[@]}"
fi
make build
echo 'Ready. In a regular Mac terminal, run ./start.sh [working-directory]'
