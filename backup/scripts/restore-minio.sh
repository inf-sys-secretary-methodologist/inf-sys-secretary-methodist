#!/bin/bash
# MinIO restore script for inf-sys-secretary-methodist
# Usage: ./restore-minio.sh [backup_file]
# If no backup file specified, uses the latest backup

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups/minio}"
BACKUP_FILE="${1:-}"
TEMP_DIR="/tmp/minio-restore"

# MinIO connection from environment
MINIO_HOST="${MINIO_HOST:-minio}"
MINIO_PORT="${MINIO_PORT:-9000}"
MINIO_ACCESS_KEY="${MINIO_ROOT_USER:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_ROOT_PASSWORD:-minioadmin}"
MINIO_BUCKET="${S3_BUCKET_NAME:-documents}"
MINIO_ALIAS="restore-target"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    log "ERROR: $1" >&2
    rm -rf "${TEMP_DIR}"
    exit 1
}

cleanup() {
    rm -rf "${TEMP_DIR}"
}
trap cleanup EXIT

# Find latest backup if not specified
if [ -z "${BACKUP_FILE}" ]; then
    log "No backup file specified, searching for latest backup..."
    BACKUP_FILE=$(ls -t "${BACKUP_DIR}"/minio_*.tar.gz 2>/dev/null | head -1)
    if [ -z "${BACKUP_FILE}" ]; then
        error "No backup files found in ${BACKUP_DIR}"
    fi
fi

# Check if backup file exists
if [ ! -f "${BACKUP_FILE}" ]; then
    error "Backup file not found: ${BACKUP_FILE}"
fi

log "Starting MinIO restore..."
log "Target: ${MINIO_HOST}:${MINIO_PORT}/${MINIO_BUCKET}"
log "Backup file: ${BACKUP_FILE}"

# Wait for MinIO to be ready
log "Waiting for MinIO to be ready..."
for i in {1..30}; do
    if curl -s "http://${MINIO_HOST}:${MINIO_PORT}/minio/health/live" > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        error "MinIO is not ready after 30 attempts"
    fi
    sleep 2
done
log "MinIO is ready"

# Confirm restore
log "WARNING: This will overwrite the current bucket data!"
log "Bucket: ${MINIO_BUCKET}"
log "Backup: ${BACKUP_FILE}"

if [ "${RESTORE_CONFIRM:-false}" != "true" ]; then
    log "Set RESTORE_CONFIRM=true to proceed with restore"
    exit 1
fi

# Configure mc alias
log "Configuring MinIO client..."
mc alias set "${MINIO_ALIAS}" "http://${MINIO_HOST}:${MINIO_PORT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}" --api S3v4 > /dev/null

# Create temp directory and extract backup
log "Extracting backup..."
mkdir -p "${TEMP_DIR}"
tar -xzf "${BACKUP_FILE}" -C "${TEMP_DIR}"

# Find extracted backup directory
BACKUP_CONTENT=$(ls "${TEMP_DIR}" | head -1)
if [ -z "${BACKUP_CONTENT}" ]; then
    error "Backup archive is empty"
fi

# Ensure bucket exists
log "Ensuring bucket exists..."
mc mb "${MINIO_ALIAS}/${MINIO_BUCKET}" 2>/dev/null || true

# Clear existing bucket content
log "Clearing existing bucket content..."
mc rm --recursive --force "${MINIO_ALIAS}/${MINIO_BUCKET}" 2>/dev/null || true

# Mirror backup to MinIO
log "Restoring files to MinIO..."
RESTORED_PATH="${TEMP_DIR}/${BACKUP_CONTENT}/${MINIO_BUCKET}"
if [ -d "${RESTORED_PATH}" ]; then
    if mc mirror --preserve "${RESTORED_PATH}" "${MINIO_ALIAS}/${MINIO_BUCKET}"; then
        FILE_COUNT=$(find "${RESTORED_PATH}" -type f | wc -l)
        log "Restored ${FILE_COUNT} files successfully"
    else
        error "Failed to restore files"
    fi
else
    log "No files to restore (bucket directory not found in backup)"
fi

log "MinIO restore completed successfully"
