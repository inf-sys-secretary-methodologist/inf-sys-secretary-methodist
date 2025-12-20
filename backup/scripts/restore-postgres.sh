#!/bin/bash
# PostgreSQL restore script for inf-sys-secretary-methodist
# Usage: ./restore-postgres.sh [backup_file]
# If no backup file specified, uses the latest backup

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
BACKUP_FILE="${1:-}"

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

# Find latest backup if not specified
if [ -z "${BACKUP_FILE}" ]; then
    log "No backup file specified, searching for latest backup..."
    BACKUP_FILE=$(ls -t "${BACKUP_DIR}"/postgres_*.sql.gz 2>/dev/null | head -1)
    if [ -z "${BACKUP_FILE}" ]; then
        error "No backup files found in ${BACKUP_DIR}"
    fi
fi

# Check if backup file exists
if [ ! -f "${BACKUP_FILE}" ]; then
    error "Backup file not found: ${BACKUP_FILE}"
fi

log "Starting PostgreSQL restore..."
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

# Confirm restore
log "WARNING: This will overwrite the current database!"
log "Database: ${DB_NAME}"
log "Backup: ${BACKUP_FILE}"

if [ "${RESTORE_CONFIRM:-false}" != "true" ]; then
    log "Set RESTORE_CONFIRM=true to proceed with restore"
    exit 1
fi

# Terminate existing connections
log "Terminating existing connections..."
PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d postgres \
    -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '${DB_NAME}' AND pid <> pg_backend_pid();" \
    > /dev/null 2>&1 || true

# Restore database
log "Restoring database..."
if gunzip -c "${BACKUP_FILE}" | PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    --single-transaction \
    > /dev/null 2>&1; then
    log "Database restored successfully"
else
    error "Failed to restore database"
fi

# Verify restore
log "Verifying restore..."
TABLE_COUNT=$(PGPASSWORD="${DB_PASSWORD}" psql \
    -h "${DB_HOST}" \
    -p "${DB_PORT}" \
    -U "${DB_USER}" \
    -d "${DB_NAME}" \
    -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';")
log "Restored database has ${TABLE_COUNT} tables"

log "PostgreSQL restore completed successfully"
