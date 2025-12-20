#!/bin/bash
# PostgreSQL backup script for inf-sys-secretary-methodist
# Usage: ./backup-postgres.sh [backup_dir]

set -euo pipefail

# Configuration
BACKUP_DIR="${1:-/backups/postgres}"
RETENTION_DAYS="${POSTGRES_BACKUP_RETENTION:-7}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/postgres_${TIMESTAMP}.sql.gz"

# Database connection from environment
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-inf_sys_db}"
DB_USER="${DB_USER:-postgres}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

error() {
    log "ERROR: $1" >&2
    exit 1
}

# Create backup directory if not exists
mkdir -p "${BACKUP_DIR}"

log "Starting PostgreSQL backup..."
log "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"
log "Backup file: ${BACKUP_FILE}"

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

    BACKUP_SIZE=$(du -h "${BACKUP_FILE}" | cut -f1)
    log "Backup created successfully: ${BACKUP_FILE} (${BACKUP_SIZE})"
else
    error "Failed to create backup"
fi

# Clean old backups
log "Cleaning backups older than ${RETENTION_DAYS} days..."
DELETED_COUNT=$(find "${BACKUP_DIR}" -name "postgres_*.sql.gz" -type f -mtime +${RETENTION_DAYS} -delete -print | wc -l)
log "Deleted ${DELETED_COUNT} old backup(s)"

# List current backups
log "Current backups:"
ls -lh "${BACKUP_DIR}"/postgres_*.sql.gz 2>/dev/null | tail -5 || log "No backups found"

log "PostgreSQL backup completed successfully"
