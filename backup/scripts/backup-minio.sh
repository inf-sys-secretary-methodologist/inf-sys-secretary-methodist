#!/bin/bash
# MinIO backup script for inf-sys-secretary-methodist
# Features: compression, encryption (GPG/age), metrics, notifications
# Usage: ./backup-minio.sh [backup_dir]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${1:-/backups/minio}"
RETENTION_DAYS="${MINIO_BACKUP_RETENTION:-7}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_PATH="${BACKUP_DIR}/${TIMESTAMP}"

# Encryption settings
ENCRYPTION_ENABLED="${BACKUP_ENCRYPTION_ENABLED:-false}"
ENCRYPTION_TYPE="${BACKUP_ENCRYPTION_TYPE:-age}"  # age or gpg
GPG_RECIPIENT="${BACKUP_GPG_RECIPIENT:-}"
AGE_PUBLIC_KEY="${BACKUP_AGE_PUBLIC_KEY:-}"

# MinIO connection from environment
MINIO_HOST="${MINIO_HOST:-minio}"
MINIO_PORT="${MINIO_PORT:-9000}"
MINIO_ACCESS_KEY="${MINIO_ROOT_USER:-minioadmin}"
MINIO_SECRET_KEY="${MINIO_ROOT_PASSWORD:-minioadmin}"
MINIO_BUCKET="${S3_BUCKET_NAME:-documents}"
MINIO_ALIAS="backup-source"

# Metrics and notifications
METRICS_ENABLED="${METRICS_ENABLED:-true}"
NOTIFY_ON_SUCCESS="${NOTIFY_ON_SUCCESS:-false}"
NOTIFY_ON_FAILURE="${NOTIFY_ON_FAILURE:-true}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    log "ERROR: $1" >&2

    # Send failure notification
    if [[ "${NOTIFY_ON_FAILURE}" == "true" ]] && [[ -x "${SCRIPT_DIR}/notify.sh" ]]; then
        "${SCRIPT_DIR}/notify.sh" error "MinIO Backup Failed" "$1" || true
    fi

    # Record failure metric
    if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
        "${SCRIPT_DIR}/metrics.sh" record_backup minio failure 0 0 || true
    fi

    exit 1
}

# Create backup directory
mkdir -p "${BACKUP_PATH}"

START_TIME=$(date +%s)

log "Starting MinIO backup..."
log "Source: ${MINIO_HOST}:${MINIO_PORT}/${MINIO_BUCKET}"
log "Backup path: ${BACKUP_PATH}"
log "Encryption: ${ENCRYPTION_ENABLED} (${ENCRYPTION_TYPE})"

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

    # Record as success with 0 size (empty backup)
    if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
        "${SCRIPT_DIR}/metrics.sh" record_backup minio success 0 0 || true
    fi

    exit 0
fi

# Mirror bucket to backup location
log "Mirroring bucket to backup location..."
if mc mirror --preserve "${MINIO_ALIAS}/${MINIO_BUCKET}" "${BACKUP_PATH}/${MINIO_BUCKET}"; then
    # Calculate backup size
    BACKUP_SIZE_HUMAN=$(du -sh "${BACKUP_PATH}" | cut -f1)
    FILE_COUNT=$(find "${BACKUP_PATH}" -type f | wc -l)
    log "Backup completed: ${BACKUP_PATH} (${BACKUP_SIZE_HUMAN}, ${FILE_COUNT} files)"
else
    error "Failed to create MinIO backup"
fi

# Create a compressed archive of the backup
log "Creating compressed archive..."
ARCHIVE_FILE="${BACKUP_DIR}/minio_${TIMESTAMP}.tar.gz"
if tar -czf "${ARCHIVE_FILE}" -C "${BACKUP_DIR}" "${TIMESTAMP}"; then
    ARCHIVE_SIZE_HUMAN=$(du -h "${ARCHIVE_FILE}" | cut -f1)
    ARCHIVE_SIZE=$(stat -f%z "${ARCHIVE_FILE}" 2>/dev/null || stat -c%s "${ARCHIVE_FILE}" 2>/dev/null || echo 0)
    log "Archive created: ${ARCHIVE_FILE} (${ARCHIVE_SIZE_HUMAN})"
    # Remove uncompressed backup
    rm -rf "${BACKUP_PATH}"
else
    error "Failed to create archive"
fi

# Apply encryption if enabled
FINAL_FILE="${ARCHIVE_FILE}"
if [[ "${ENCRYPTION_ENABLED}" == "true" ]]; then
    log "Encrypting backup..."

    case "${ENCRYPTION_TYPE}" in
        age)
            if [[ -z "${AGE_PUBLIC_KEY}" ]]; then
                error "BACKUP_AGE_PUBLIC_KEY is required for age encryption"
            fi

            if ! command -v age &> /dev/null; then
                error "age is not installed but encryption is enabled"
            fi

            ENCRYPTED_FILE="${ARCHIVE_FILE}.age"
            if age -r "${AGE_PUBLIC_KEY}" -o "${ENCRYPTED_FILE}" "${ARCHIVE_FILE}"; then
                rm -f "${ARCHIVE_FILE}"
                FINAL_FILE="${ENCRYPTED_FILE}"
                log "Encrypted with age: ${FINAL_FILE}"
            else
                error "Failed to encrypt with age"
            fi
            ;;

        gpg)
            if [[ -z "${GPG_RECIPIENT}" ]]; then
                error "BACKUP_GPG_RECIPIENT is required for GPG encryption"
            fi

            if ! command -v gpg &> /dev/null; then
                error "gpg is not installed but encryption is enabled"
            fi

            ENCRYPTED_FILE="${ARCHIVE_FILE}.gpg"
            if gpg --encrypt --recipient "${GPG_RECIPIENT}" --output "${ENCRYPTED_FILE}" "${ARCHIVE_FILE}"; then
                rm -f "${ARCHIVE_FILE}"
                FINAL_FILE="${ENCRYPTED_FILE}"
                log "Encrypted with GPG: ${FINAL_FILE}"
            else
                error "Failed to encrypt with GPG"
            fi
            ;;

        *)
            error "Unknown encryption type: ${ENCRYPTION_TYPE}"
            ;;
    esac

    # Update size after encryption
    ARCHIVE_SIZE=$(stat -f%z "${FINAL_FILE}" 2>/dev/null || stat -c%s "${FINAL_FILE}" 2>/dev/null || echo 0)
fi

# Clean old backups
log "Cleaning backups older than ${RETENTION_DAYS} days..."
DELETED_COUNT=$(find "${BACKUP_DIR}" -name "minio_*.tar.gz*" -type f -mtime +${RETENTION_DAYS} -delete -print | wc -l)
log "Deleted ${DELETED_COUNT} old backup(s)"

# List current backups
log "Current backups:"
ls -lh "${BACKUP_DIR}"/minio_*.tar.gz* 2>/dev/null | tail -5 || log "No backups found"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log "MinIO backup completed successfully in ${DURATION}s"

# Record success metric
if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
    "${SCRIPT_DIR}/metrics.sh" record_backup minio success "${DURATION}" "${ARCHIVE_SIZE}" || true
fi

# Send success notification if enabled
if [[ "${NOTIFY_ON_SUCCESS}" == "true" ]] && [[ -x "${SCRIPT_DIR}/notify.sh" ]]; then
    "${SCRIPT_DIR}/notify.sh" success "MinIO Backup Complete" \
        "Bucket: ${MINIO_BUCKET}
Files: ${FILE_COUNT}
Size: ${ARCHIVE_SIZE_HUMAN}
Duration: ${DURATION}s
Archive: $(basename "${FINAL_FILE}")" || true
fi
