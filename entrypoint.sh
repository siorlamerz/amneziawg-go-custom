#!/bin/sh
set -e

CONF="${AWG_CONFIG:-/etc/amneziawg/awg0.conf}"

if [ ! -f "$CONF" ]; then
    echo "ERROR: config not found at $CONF" >&2
    echo "Mount your config: -v /path/to/awg0.conf:$CONF:ro" >&2
    exit 1
fi

cleanup() {
    echo "Shutting down interface..."
    awg-quick down "$CONF" 2>/dev/null || true
    exit 0
}
trap cleanup INT TERM

awg-quick up "$CONF"

# keep the container running and forward signals
while true; do
    sleep 3600 &
    wait $!
done
