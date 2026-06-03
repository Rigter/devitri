#!/bin/sh
set -e

DATA_DIR="${DEVITRI_DATA_DIR:-/data}"
VAULTS_DIR="${DEVITRI_VAULTS_DIR:-/vaults}"

echo "[devitri-entrypoint] preparing ${DATA_DIR} and ${VAULTS_DIR}"

mkdir -p "${DATA_DIR}" "${VAULTS_DIR}"

DEVITRI_UID="$(id -u devitri)"
DEVITRI_GID="$(id -g devitri)"
chown -R "${DEVITRI_UID}:${DEVITRI_GID}" "${DATA_DIR}" "${VAULTS_DIR}"

# Verify the app user can write (fails fast with a clear message in docker logs)
su-exec "${DEVITRI_UID}:${DEVITRI_GID}" sh -c "touch \"${DATA_DIR}/.write-test\" && rm \"${DATA_DIR}/.write-test\""

echo "[devitri-entrypoint] starting server as uid=${DEVITRI_UID} gid=${DEVITRI_GID}"
exec su-exec "${DEVITRI_UID}:${DEVITRI_GID}" ./devitri
