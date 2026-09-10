#!/bin/bash
set -euo pipefail
cd "$(dirname "$0")"
if ! command -v brew >/dev/null; then echo 'Install Homebrew first: https://brew.sh'; exit 1; fi
if ! xcode-select -p >/dev/null 2>&1; then echo 'Run xcode-select --install first.'; exit 1; fi
missing=()
for package in cmake libuv json-c libwebsockets openssl@3 herdr; do
  if ! brew list --versions "$package" >/dev/null 2>&1; then missing+=("$package"); fi
done
if [ "${#missing[@]}" -gt 0 ]; then brew install "${missing[@]}"; fi
brew_prefix=$(brew --prefix)
"$brew_prefix/bin/cmake" -S vendor/ttyd -B build -DCMAKE_BUILD_TYPE=Release -DCMAKE_PREFIX_PATH="$brew_prefix" -DOPENSSL_ROOT_DIR="$(brew --prefix openssl@3)"
"$brew_prefix/bin/cmake" --build build -j 4
echo 'Ready. Run ./start.sh [working-directory]'
