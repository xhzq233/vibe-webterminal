#!/bin/bash
set -euo pipefail
bundle_dir=$(cd "$(dirname "$0")" && pwd)
work_dir=${1:-$HOME}
if [ ! -x "$bundle_dir/build/ttyd" ]; then echo 'Run ./setup.sh first.'; exit 1; fi
state_dir=${VIBE_STATE_DIR:-$HOME/.local/share/vibe-webterminal}
mkdir -p "$state_dir"
chmod 700 "$state_dir"
if [ ! -f "$state_dir/credential" ]; then
  (umask 077; printf 'terminal:%s' "$(openssl rand -hex 18)" > "$state_dir/credential")
fi
iface=$(/sbin/route -n get default | /usr/bin/awk '/interface:/ {print $2}')
ip=${BIND_IP:-$(/usr/sbin/ipconfig getifaddr "$iface")}
port=${PORT:-7681}
echo "Open: http://$ip:$port"
echo "Credentials: $state_dir/credential (username:password)"
echo 'Stop: Ctrl+C in this Mac terminal. No auto-start is installed.'
exec "$bundle_dir/build/ttyd" -i "$ip" -p "$port" -I "$bundle_dir/web/index.html" -W -O -c "$(cat "$state_dir/credential")" -w "$work_dir" /bin/zsh -l
