#!/bin/sh
set -e

# Named volumes are often root-owned; the app runs as devitri and must write SQLite + vault files.
mkdir -p /data /vaults
chown -R devitri:devitri /data /vaults

exec su-exec devitri:devitri ./devitri
