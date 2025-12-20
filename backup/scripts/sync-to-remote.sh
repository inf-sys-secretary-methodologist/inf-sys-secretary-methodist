#!/bin/bash
# Sync backups to remote S3-compatible storage
# Supports: AWS S3, Backblaze B2, Yandex Object Storage, MinIO, etc.
# Usage: ./sync-to-remote.sh

set -euo pipefail

# Remote S3 configuration
REMOTE_S3_ENDPOINT="${REMOTE_S3_ENDPOINT:-}"
REMOTE_S3_ACCESS_KEY="${REMOTE_S3_ACCESS_KEY:-}"
REMOTE_S3_SECRET_KEY="${REMOTE_S3_SECRET_KEY:-}"
REMOTE_S3_BUCKET="${REMOTE_S3_BUCKET:-}"
REMOTE_S3_REGION="${REMOTE_S3_REGION:-us-east-1}"
REMOTE_S3_PATH="${REMOTE_S3_PATH:-backups}"

# Sync settings
SYNC_ENABLED="${REMOTE_SYNC_ENABLED:-false}"
BACKUP_DIR="${BACKUP_BASE_DIR:-/backups}"
REMOTE_ALIAS="remote-backup"

# Server identifier for multi-server setups
SERVER_ID="${SERVER_ID:-$(hostname)}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [SYNC] $1"
}

error() {
    log "ERROR: $1" >&2
    exit 1
}

# Check if sync is enabled
if [ "${SYNC_ENABLED}" != "true" ]; then
    log "Remote sync is disabled (REMOTE_SYNC_ENABLED != true)"
    exit 0
fi

# Validate required variables
if [ -z "${REMOTE_S3_ENDPOINT}" ]; then
    error "REMOTE_S3_ENDPOINT is not set"
fi

if [ -z "${REMOTE_S3_ACCESS_KEY}" ]; then
    error "REMOTE_S3_ACCESS_KEY is not set"
fi

if [ -z "${REMOTE_S3_SECRET_KEY}" ]; then
    error "REMOTE_S3_SECRET_KEY is not set"
fi

if [ -z "${REMOTE_S3_BUCKET}" ]; then
    error "REMOTE_S3_BUCKET is not set"
fi

log "Starting remote sync..."
log "Endpoint: ${REMOTE_S3_ENDPOINT}"
log "Bucket: ${REMOTE_S3_BUCKET}"
log "Path: ${REMOTE_S3_PATH}/${SERVER_ID}/"

# Configure mc alias for remote storage
log "Configuring remote storage connection..."
mc alias set "${REMOTE_ALIAS}" \
    "${REMOTE_S3_ENDPOINT}" \
    "${REMOTE_S3_ACCESS_KEY}" \
    "${REMOTE_S3_SECRET_KEY}" \
    --api S3v4 > /dev/null

# Test connection
log "Testing connection..."
if ! mc ls "${REMOTE_ALIAS}" > /dev/null 2>&1; then
    error "Failed to connect to remote storage"
fi

# Ensure bucket exists
if ! mc ls "${REMOTE_ALIAS}/${REMOTE_S3_BUCKET}" > /dev/null 2>&1; then
    log "Creating bucket ${REMOTE_S3_BUCKET}..."
    mc mb "${REMOTE_ALIAS}/${REMOTE_S3_BUCKET}" || error "Failed to create bucket"
fi

# Sync PostgreSQL backups
REMOTE_PG_PATH="${REMOTE_ALIAS}/${REMOTE_S3_BUCKET}/${REMOTE_S3_PATH}/${SERVER_ID}/postgres"
if [ -d "${BACKUP_DIR}/postgres" ] && [ "$(ls -A ${BACKUP_DIR}/postgres 2>/dev/null)" ]; then
    log "Syncing PostgreSQL backups..."
    if mc mirror --preserve --overwrite "${BACKUP_DIR}/postgres" "${REMOTE_PG_PATH}"; then
        PG_COUNT=$(ls -1 "${BACKUP_DIR}/postgres"/*.sql.gz 2>/dev/null | wc -l)
        log "PostgreSQL sync complete: ${PG_COUNT} backup(s)"
    else
        log "WARNING: PostgreSQL sync failed"
    fi
else
    log "No PostgreSQL backups to sync"
fi

# Sync MinIO backups
REMOTE_MINIO_PATH="${REMOTE_ALIAS}/${REMOTE_S3_BUCKET}/${REMOTE_S3_PATH}/${SERVER_ID}/minio"
if [ -d "${BACKUP_DIR}/minio" ] && [ "$(ls -A ${BACKUP_DIR}/minio 2>/dev/null)" ]; then
    log "Syncing MinIO backups..."
    if mc mirror --preserve --overwrite "${BACKUP_DIR}/minio" "${REMOTE_MINIO_PATH}"; then
        MINIO_COUNT=$(ls -1 "${BACKUP_DIR}/minio"/*.tar.gz 2>/dev/null | wc -l)
        log "MinIO sync complete: ${MINIO_COUNT} backup(s)"
    else
        log "WARNING: MinIO sync failed"
    fi
else
    log "No MinIO backups to sync"
fi

# Show remote backup summary
log "Remote backup summary:"
log "PostgreSQL backups:"
mc ls "${REMOTE_PG_PATH}" 2>/dev/null | tail -3 || echo "  No backups"
log "MinIO backups:"
mc ls "${REMOTE_MINIO_PATH}" 2>/dev/null | tail -3 || echo "  No backups"

log "Remote sync completed successfully"
