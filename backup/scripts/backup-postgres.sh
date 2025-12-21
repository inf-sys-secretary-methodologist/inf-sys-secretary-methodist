#!/bin/bash
# PostgreSQL backup script for inf-sys-secretary-methodist
# Features: compression, encryption (GPG/age), metrics, notifications
# Usage: ./backup-postgres.sh [backup_dir]

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${1:-/backups/postgres}"
RETENTION_DAYS="${POSTGRES_BACKUP_RETENTION:-7}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/postgres_${TIMESTAMP}.sql.gz"

# Encryption settings
ENCRYPTION_ENABLED="${BACKUP_ENCRYPTION_ENABLED:-false}"
ENCRYPTION_TYPE="${BACKUP_ENCRYPTION_TYPE:-age}"  # age or gpg
GPG_RECIPIENT="${BACKUP_GPG_RECIPIENT:-}"
AGE_PUBLIC_KEY="${BACKUP_AGE_PUBLIC_KEY:-}"

# Database connection from environment
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-inf_sys_db}"
DB_USER="${DB_USER:-postgres}"

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
        "${SCRIPT_DIR}/notify.sh" error "PostgreSQL Backup Failed" "$1" || true
    fi

    # Record failure metric
    if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
        "${SCRIPT_DIR}/metrics.sh" record_backup postgres failure 0 0 || true
    fi

    exit 1
}

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

START_TIME=$(date +%s)

log "Starting PostgreSQL backup..."
log "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
log "Backup file: ${BACKUP_FILE}"
log "Encryption: ${ENCRYPTION_ENABLED} (${ENCRYPTION_TYPE})"

# Wait for PostgreSQL to be ready
log "Waiting for PostgreSQL to be ready..."
for i in {1..30}; do
    if PGPASSWORD="${DB_PASSWORD}" pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" > /dev/null 2>&1; then
        break
    fi
    if [ $i -eq 30 ]; then
        error "PostgreSQL is not ready after 30 attempts"
    fi
    sleep 2
done
log "PostgreSQL is ready"

# Create backup
log "Creating backup..."
if PGPASSWORD="${DB_PASSWORD}" pg_dump \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --format=plain \
    --no-owner \
    --no-acl \
    --clean \
    --if-exists \
    | gzip > "${BACKUP_FILE}"; then

    BACKUP_SIZE=$(stat -f%z "${BACKUP_FILE}" 2>/dev/null || stat -c%s "${BACKUP_FILE}" 2>/dev/null || echo 0)
    BACKUP_SIZE_HUMAN=$(du -h "${BACKUP_FILE}" | cut -f1)
    log "Backup created successfully: ${BACKUP_FILE} (${BACKUP_SIZE_HUMAN})"
else
    error "Failed to create backup"
fi

# Apply encryption if enabled
if [[ "${ENCRYPTION_ENABLED}" == "true" ]]; then
    log "Encrypting backup..."
    ENCRYPTED_FILE="${BACKUP_FILE}"

    case "${ENCRYPTION_TYPE}" in
        age)
            if [[ -z "${AGE_PUBLIC_KEY}" ]]; then
                error "BACKUP_AGE_PUBLIC_KEY is required for age encryption"
            fi

            if ! command -v age &> /dev/null; then
                error "age is not installed but encryption is enabled"
            fi

            ENCRYPTED_FILE="${BACKUP_FILE}.age"
            if age -r "${AGE_PUBLIC_KEY}" -o "${ENCRYPTED_FILE}" "${BACKUP_FILE}"; then
                rm -f "${BACKUP_FILE}"
                BACKUP_FILE="${ENCRYPTED_FILE}"
                log "Encrypted with age: ${BACKUP_FILE}"
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

            ENCRYPTED_FILE="${BACKUP_FILE}.gpg"
            if gpg --encrypt --recipient "${GPG_RECIPIENT}" --output "${ENCRYPTED_FILE}" "${BACKUP_FILE}"; then
                rm -f "${BACKUP_FILE}"
                BACKUP_FILE="${ENCRYPTED_FILE}"
                log "Encrypted with GPG: ${BACKUP_FILE}"
            else
                error "Failed to encrypt with GPG"
            fi
            ;;

        *)
            error "Unknown encryption type: ${ENCRYPTION_TYPE}"
            ;;
    esac

    # Update size after encryption
    BACKUP_SIZE=$(stat -f%z "${BACKUP_FILE}" 2>/dev/null || stat -c%s "${BACKUP_FILE}" 2>/dev/null || echo 0)
fi

# Clean old backups
log "Cleaning backups older than ${RETENTION_DAYS} days..."
DELETED_COUNT=$(find "${BACKUP_DIR}" -name "postgres_*.sql.gz*" -type f -mtime +${RETENTION_DAYS} -delete -print | wc -l)
log "Deleted ${DELETED_COUNT} old backup(s)"

# List current backups
log "Current backups:"
ls -lh "${BACKUP_DIR}"/postgres_*.sql.gz* 2>/dev/null | tail -5 || log "No backups found"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log "PostgreSQL backup completed successfully in ${DURATION}s"

# Record success metric
if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
    "${SCRIPT_DIR}/metrics.sh" record_backup postgres success "${DURATION}" "${BACKUP_SIZE}" || true
fi

# Send success notification if enabled
if [[ "${NOTIFY_ON_SUCCESS}" == "true" ]] && [[ -x "${SCRIPT_DIR}/notify.sh" ]]; then
    "${SCRIPT_DIR}/notify.sh" success "PostgreSQL Backup Complete" \
        "Database: ${DB_NAME}
Size: ${BACKUP_SIZE_HUMAN}
Duration: ${DURATION}s
File: $(basename "${BACKUP_FILE}")" || true
fi
