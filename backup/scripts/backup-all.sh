#!/bin/bash
# Combined backup script for inf-sys-secretary-methodist
# Backs up PostgreSQL and MinIO
# Usage: ./backup-all.sh

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_BASE_DIR="${BACKUP_BASE_DIR:-/backups}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log "=========================================="
log "Starting full system backup"
log "=========================================="

# Backup PostgreSQL
log ""
log "--- PostgreSQL Backup ---"
if "${SCRIPT_DIR}/backup-postgres.sh" "${BACKUP_BASE_DIR}/postgres"; then
    log "PostgreSQL backup: SUCCESS"
else
    log "PostgreSQL backup: FAILED"
    exit 1
fi

# Backup MinIO
log ""
log "--- MinIO Backup ---"
if "${SCRIPT_DIR}/backup-minio.sh" "${BACKUP_BASE_DIR}/minio"; then
    log "MinIO backup: SUCCESS"
else
    log "MinIO backup: FAILED"
    exit 1
fi

log ""
log "=========================================="
log "Local backup completed successfully"
log "=========================================="

# Show backup summary
log ""
log "Backup Summary:"
log "PostgreSQL backups:"
ls -lh "${BACKUP_BASE_DIR}/postgres/"postgres_*.sql.gz 2>/dev/null | tail -3 || echo "  No backups"
log "MinIO backups:"
ls -lh "${BACKUP_BASE_DIR}/minio/"minio_*.tar.gz 2>/dev/null | tail -3 || echo "  No backups"

# Sync to remote storage (if enabled)
log ""
log "--- Remote Sync ---"
if "${SCRIPT_DIR}/sync-to-remote.sh"; then
    log "Remote sync: SUCCESS (or disabled)"
else
    log "Remote sync: FAILED (non-critical, local backup is safe)"
fi

log ""
log "=========================================="
log "Full system backup completed"
log "=========================================="
