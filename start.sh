#!/bin/bash
set -euo pipefail
bundle_dir=$(cd "$(dirname "$0")" && pwd)
if [ -n "${HERDR_ENV:-}" ]; then
  echo 'Start from a regular Mac terminal or service, outside a Herdr pane. Existing Herdr sessions can still be attached.'
  exit 1
fi
work_dir=${1:-$HOME}
session_name=${VIBE_SESSION:-vibe-webterminal}
herdr_bin=$(command -v herdr) || { echo 'Herdr is missing. Run ./setup.sh first.'; exit 1; }
ttyd_bin=$(command -v ttyd) || { echo 'ttyd is missing. Run ./setup.sh first.'; exit 1; }
if [ ! -x "$bundle_dir/bin/herdr-tty" ]; then echo 'Run ./setup.sh first.'; exit 1; fi
state_dir=${VIBE_STATE_DIR:-$HOME/.local/share/vibe-webterminal}
mkdir -p "$state_dir"
chmod 700 "$state_dir"
if [ ! -f "$state_dir/credential" ]; then
  (umask 077; printf 'terminal:%s' "$(openssl rand -hex 18)" > "$state_dir/credential")
fi
credential=$(cat "$state_dir/credential")
export HERDR_TTY_USERNAME="${credential%%:*}"
export HERDR_TTY_PASSWORD="${credential#*:}"
if [ -n "${BIND_IP:-}" ]; then
  ip=$BIND_IP
else
  iface=$(/sbin/route -n get default | /usr/bin/awk '/interface:/ {print $2}')
  ip=$(/usr/sbin/ipconfig getifaddr "$iface") || { echo 'Set BIND_IP to a reachable interface IPv4 address.'; exit 1; }
fi
port=${PORT:-7681}
key_file=${VIBE_SESSION_KEY_FILE:-$state_dir/session-key}
echo "Open: http://$ip:$port"
echo "Login credentials: $state_dir/credential (username:password)"
echo "Herdr session: $session_name (reconnects reuse this session)"
echo 'Stop web access: Ctrl+C here. Herdr and its tasks keep running.'
exec "$bundle_dir/bin/herdr-tty" --listen "$ip:$port" --auth form \
  --session-key-file "$key_file" --session "$session_name" --cwd "$work_dir" \
  --no-open --ttyd "$ttyd_bin" --herdr "$herdr_bin"
