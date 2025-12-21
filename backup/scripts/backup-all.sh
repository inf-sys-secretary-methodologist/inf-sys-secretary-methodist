#!/bin/bash
# Combined backup script for inf-sys-secretary-methodist
# Backs up PostgreSQL and MinIO with notifications and metrics
# Usage: ./backup-all.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_BASE_DIR="${BACKUP_BASE_DIR:-/backups}"

# Notification settings
NOTIFY_ON_SUCCESS="${NOTIFY_ON_SUCCESS:-false}"
NOTIFY_ON_FAILURE="${NOTIFY_ON_FAILURE:-true}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Track overall status
PG_STATUS="pending"
MINIO_STATUS="pending"
SYNC_STATUS="pending"
OVERALL_STATUS="success"
ERRORS=""

log "=========================================="
log "Starting full system backup"
log "=========================================="

START_TIME=$(date +%s)

# Backup PostgreSQL
log ""
log "--- PostgreSQL Backup ---"
if "${SCRIPT_DIR}/backup-postgres.sh" "${BACKUP_BASE_DIR}/postgres"; then
    log "PostgreSQL backup: SUCCESS"
    PG_STATUS="success"
else
    log "PostgreSQL backup: FAILED"
    PG_STATUS="failed"
    OVERALL_STATUS="failed"
    ERRORS="${ERRORS}PostgreSQL backup failed\n"
fi

# Backup MinIO
log ""
log "--- MinIO Backup ---"
if "${SCRIPT_DIR}/backup-minio.sh" "${BACKUP_BASE_DIR}/minio"; then
    log "MinIO backup: SUCCESS"
    MINIO_STATUS="success"
else
    log "MinIO backup: FAILED"
    MINIO_STATUS="failed"
    OVERALL_STATUS="failed"
    ERRORS="${ERRORS}MinIO backup failed\n"
fi

log ""
log "=========================================="
log "Local backup completed"
log "=========================================="

# Show backup summary
log ""
log "Backup Summary:"
log "PostgreSQL backups:"
ls -lh "${BACKUP_BASE_DIR}/postgres/"postgres_*.sql.gz* 2>/dev/null | tail -3 || echo "  No backups"
log "MinIO backups:"
ls -lh "${BACKUP_BASE_DIR}/minio/"minio_*.tar.gz* 2>/dev/null | tail -3 || echo "  No backups"

# Sync to remote storage (if enabled)
log ""
log "--- Remote Sync ---"
if "${SCRIPT_DIR}/sync-to-remote.sh"; then
    log "Remote sync: SUCCESS (or disabled)"
    SYNC_STATUS="success"
else
    log "Remote sync: FAILED (non-critical, local backup is safe)"
    SYNC_STATUS="failed"
    # Don't fail overall if just sync failed - local backup is still valid
    ERRORS="${ERRORS}Remote sync failed (non-critical)\n"
fi

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))

log ""
log "=========================================="
log "Full system backup completed in ${DURATION}s"
log "Status: ${OVERALL_STATUS^^}"
log "=========================================="

# Send summary notification
if [[ -x "${SCRIPT_DIR}/notify.sh" ]]; then
    SUMMARY="PostgreSQL: ${PG_STATUS}
MinIO: ${MINIO_STATUS}
Remote Sync: ${SYNC_STATUS}
Duration: ${DURATION}s"

    if [[ "${OVERALL_STATUS}" == "failed" ]]; then
        if [[ "${NOTIFY_ON_FAILURE}" == "true" ]]; then
            "${SCRIPT_DIR}/notify.sh" error "Backup Failed" "${SUMMARY}

Errors:
${ERRORS}" || true
        fi
    else
        if [[ "${NOTIFY_ON_SUCCESS}" == "true" ]]; then
            "${SCRIPT_DIR}/notify.sh" success "Backup Complete" "${SUMMARY}" || true
        fi
    fi
fi

# Exit with appropriate status
if [[ "${OVERALL_STATUS}" == "failed" ]]; then
    exit 1
fi
