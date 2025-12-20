#!/bin/bash
# MinIO backup script for inf-sys-secretary-methodist
# Uses mc (MinIO Client) to mirror data to backup location
# Usage: ./backup-minio.sh [backup_dir]

set -euo pipefail

# Configuration
BACKUP_DIR="${1:-/backups/minio}"
RETENTION_DAYS="${MINIO_BACKUP_RETENTION:-7}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_PATH="${BACKUP_DIR}/${TIMESTAMP}"

# MinIO connection from environment
MINIO_HOST="${MINIO_HOST:-minio}"
MINIO_PORT="${MINIO_PORT:-9000}"
MINIO_ACCESS_KEY="${MINIO_ROOT_USER:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_ROOT_PASSWORD:-minioadmin}"
MINIO_BUCKET="${S3_BUCKET_NAME:-documents}"
MINIO_ALIAS="backup-source"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    log "ERROR: $1" >&2
    exit 1
}

# Create backup directory
mkdir -p "${BACKUP_PATH}"

log "Starting MinIO backup..."
log "Source: ${MINIO_HOST}:${MINIO_PORT}/${MINIO_BUCKET}"
log "Backup path: ${BACKUP_PATH}"

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

# Configure mc alias
log "Configuring MinIO client..."
mc alias set "${MINIO_ALIAS}" "http://${MINIO_HOST}:${MINIO_PORT}" "${MINIO_ACCESS_KEY}" "${MINIO_SECRET_KEY}" --api S3v4 > /dev/null

# Check if bucket exists
if ! mc ls "${MINIO_ALIAS}/${MINIO_BUCKET}" > /dev/null 2>&1; then
    log "Bucket ${MINIO_BUCKET} does not exist, skipping backup"
    exit 0
fi

# Mirror bucket to backup location
log "Mirroring bucket to backup location..."
if mc mirror --preserve "${MINIO_ALIAS}/${MINIO_BUCKET}" "${BACKUP_PATH}/${MINIO_BUCKET}"; then
    # Calculate backup size
    BACKUP_SIZE=$(du -sh "${BACKUP_PATH}" | cut -f1)
    FILE_COUNT=$(find "${BACKUP_PATH}" -type f | wc -l)
    log "Backup completed: ${BACKUP_PATH} (${BACKUP_SIZE}, ${FILE_COUNT} files)"
else
    error "Failed to create MinIO backup"
fi

# Create a compressed archive of the backup
log "Creating compressed archive..."
ARCHIVE_FILE="${BACKUP_DIR}/minio_${TIMESTAMP}.tar.gz"
if tar -czf "${ARCHIVE_FILE}" -C "${BACKUP_DIR}" "${TIMESTAMP}"; then
    ARCHIVE_SIZE=$(du -h "${ARCHIVE_FILE}" | cut -f1)
    log "Archive created: ${ARCHIVE_FILE} (${ARCHIVE_SIZE})"
    # Remove uncompressed backup
    rm -rf "${BACKUP_PATH}"
else
    error "Failed to create archive"
fi

# Clean old backups
log "Cleaning backups older than ${RETENTION_DAYS} days..."
DELETED_COUNT=$(find "${BACKUP_DIR}" -name "minio_*.tar.gz" -type f -mtime +${RETENTION_DAYS} -delete -print | wc -l)
log "Deleted ${DELETED_COUNT} old backup(s)"

# List current backups
log "Current backups:"
ls -lh "${BACKUP_DIR}"/minio_*.tar.gz 2>/dev/null | tail -5 || log "No backups found"

log "MinIO backup completed successfully"
