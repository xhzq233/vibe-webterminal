#!/usr/bin/env python3
"""Update a DNS-only A record and rebind the local Herdr web gateway to en0."""
import ipaddress
import json
import pathlib
import subprocess
import sys
import time
import urllib.error
import urllib.parse
import urllib.request


def api(config, method, suffix, payload=None):
    token = pathlib.Path(config['token_file']).expanduser().read_text().strip()
    request = urllib.request.Request(
        'https://api.cloudflare.com/client/v4/' + suffix,
        data=None if payload is None else json.dumps(payload).encode(),
        method=method,
        headers={'Authorization': 'Bearer ' + token, 'Content-Type': 'application/json'},
    )
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            result = json.load(response)
    except urllib.error.HTTPError as error:
        raise RuntimeError(f'Cloudflare request failed: HTTP {error.code}') from None
    if not result.get('success'):
        raise RuntimeError('Cloudflare rejected the request')
    return result['result']


def main():
    config_path = pathlib.Path(sys.argv[1]).expanduser()
    config = json.loads(config_path.read_text())
    current = subprocess.run(['/usr/sbin/ipconfig', 'getifaddr', 'en0'], capture_output=True, text=True)
    if current.returncode:
        print('en0 has no IPv4 address; leaving DNS unchanged')
        return
    address = str(ipaddress.IPv4Address(current.stdout.strip()))
    # Inspect the actual listener rather than a cached IP, so manual restarts
    # and failed attempts are recovered on the next scheduled run.
    listeners = subprocess.run(['/usr/sbin/lsof', '-nP', '-iTCP:7681', '-sTCP:LISTEN'], capture_output=True, text=True).stdout
    if not any('herdr-tty' in line and address + ':7681' in line for line in listeners.splitlines()):
        gateway = str(pathlib.Path(config.get('gateway', '~/.local/bin/herdr-web')).expanduser())
        subprocess.run([gateway, 'stop'], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        subprocess.run([gateway, 'start'], check=True)
        for attempt in range(10):
            try:
                with urllib.request.urlopen('http://' + address + ':7681/', timeout=2):
                    break
            except urllib.error.URLError:
                if attempt == 9:
                    raise RuntimeError('Web gateway did not start; leaving DNS unchanged') from None
                time.sleep(0.5)
        print('Rebound Web gateway to ' + address, flush=True)
    suffix = 'zones/' + config['zone_id'] + '/dns_records'
    records = api(config, 'GET', suffix + '?' + urllib.parse.urlencode({'type': 'A', 'name': config['hostname']}))
    if len(records) > 1:
        raise RuntimeError('Multiple A records exist; choose a hostname with a single A record')
    payload = {'type': 'A', 'name': config['hostname'], 'content': address, 'ttl': 60, 'proxied': False}
    if records:
        record = records[0]
        if record['content'] == address and not record.get('proxied') and record['ttl'] == 60:
            print(config['hostname'] + ' already points to ' + address)
            return
        api(config, 'PATCH', suffix + '/' + record['id'], payload)
    else:
        api(config, 'POST', suffix, payload)
    print(config['hostname'] + ' updated to ' + address + ' (DNS only)')


if __name__ == '__main__':
    try:
        main()
    except Exception as error:
        print(str(error), file=sys.stderr)
        sys.exit(1)
