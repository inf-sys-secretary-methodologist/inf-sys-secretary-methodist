#!/bin/bash
# PostgreSQL incremental backup with WAL archiving (PITR support)
# This script creates base backups and manages WAL archive
# Usage: ./backup-postgres-incremental.sh [base|wal-status|restore-pitr]
#
# Prerequisites:
#   PostgreSQL must be configured for WAL archiving:
#   - wal_level = replica
#   - archive_mode = on
#   - archive_command = 'cp %p /backups/postgres/wal/%f'

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKUP_DIR="${BACKUP_BASE_DIR:-/backups}/postgres"
WAL_DIR="${BACKUP_DIR}/wal"
BASE_DIR="${BACKUP_DIR}/base"
RETENTION_DAYS="${POSTGRES_BACKUP_RETENTION:-7}"

# Database connection
DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-inf_sys_db}"
DB_USER="${DB_USER:-postgres}"

# Metrics and notifications
METRICS_ENABLED="${METRICS_ENABLED:-true}"
NOTIFY_ON_FAILURE="${NOTIFY_ON_FAILURE:-true}"

log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] [PITR] $1"
}

error() {
    log "ERROR: $1" >&2
    exit 1
}

check_wal_archiving() {
    log "Checking WAL archiving configuration..."

    # Check if wal_level is set correctly
    local wal_level
    wal_level=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "SHOW wal_level" | tr -d ' ')

    if [[ "${wal_level}" != "replica" ]] && [[ "${wal_level}" != "logical" ]]; then
        error "WAL level is '${wal_level}', must be 'replica' or 'logical' for PITR"
    fi

    # Check if archive_mode is on
    local archive_mode
    archive_mode=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "SHOW archive_mode" | tr -d ' ')

    if [[ "${archive_mode}" != "on" ]]; then
        error "archive_mode is '${archive_mode}', must be 'on' for PITR"
    fi

    log "WAL archiving is properly configured (wal_level=${wal_level}, archive_mode=${archive_mode})"
}

create_base_backup() {
    log "Creating base backup with pg_basebackup..."

    mkdir -p "${BASE_DIR}" "${WAL_DIR}"

    local TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
    local BACKUP_PATH="${BASE_DIR}/base_${TIMESTAMP}"

    # Wait for PostgreSQL
    log "Waiting for PostgreSQL..."
    for i in {1..30}; do
        if PGPASSWORD="${DB_PASSWORD}" pg_isready -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" > /dev/null 2>&1; then
            break
        fi
        if [ $i -eq 30 ]; then
            error "PostgreSQL is not ready after 30 attempts"
        fi
        sleep 2
    done

    check_wal_archiving

    START_TIME=$(date +%s)

    # Create base backup
    if PGPASSWORD="${DB_PASSWORD}" pg_basebackup \
        -h "${DB_HOST}" \
        -p "${DB_PORT}" \
        -U "${DB_USER}" \
        -D "${BACKUP_PATH}" \
        -Ft \
        -z \
        -Xs \
        -P \
        -c fast \
        -l "base_${TIMESTAMP}"; then

        END_TIME=$(date +%s)
        DURATION=$((END_TIME - START_TIME))

        # Calculate size
        BACKUP_SIZE=$(du -sb "${BACKUP_PATH}" | cut -f1)
        BACKUP_SIZE_HUMAN=$(du -sh "${BACKUP_PATH}" | cut -f1)

        log "Base backup created: ${BACKUP_PATH} (${BACKUP_SIZE_HUMAN}) in ${DURATION}s"

        # Save backup info
        cat > "${BACKUP_PATH}/backup_info.json" << EOF
{
    "timestamp": "${TIMESTAMP}",
    "type": "base",
    "database": "${DB_NAME}",
    "size_bytes": ${BACKUP_SIZE},
    "duration_seconds": ${DURATION},
    "created_at": "$(date -Iseconds)"
}
EOF

        # Record metrics
        if [[ "${METRICS_ENABLED}" == "true" ]] && [[ -x "${SCRIPT_DIR}/metrics.sh" ]]; then
            "${SCRIPT_DIR}/metrics.sh" record_backup postgres success "${DURATION}" "${BACKUP_SIZE}" || true
        fi

        # Clean old base backups
        log "Cleaning base backups older than ${RETENTION_DAYS} days..."
        find "${BASE_DIR}" -mindepth 1 -maxdepth 1 -type d -mtime +${RETENTION_DAYS} -exec rm -rf {} \;

        # Clean old WAL files (keep files newer than oldest base backup)
        local oldest_base=$(ls -1t "${BASE_DIR}" | tail -1)
        if [[ -n "${oldest_base}" ]] && [[ -d "${BASE_DIR}/${oldest_base}" ]]; then
            local oldest_time=$(stat -c %Y "${BASE_DIR}/${oldest_base}" 2>/dev/null || stat -f %m "${BASE_DIR}/${oldest_base}" 2>/dev/null)
            log "Cleaning WAL files older than oldest base backup..."
            find "${WAL_DIR}" -type f -name "*.gz" ! -newermt "@${oldest_time}" -delete 2>/dev/null || true
        fi

        log "Base backup completed successfully"
    else
        error "Failed to create base backup"
    fi
}

show_wal_status() {
    log "WAL Archive Status"
    log "=================="

    # Check configuration
    check_wal_archiving || true

    # Show WAL directory stats
    if [[ -d "${WAL_DIR}" ]]; then
        local wal_count=$(find "${WAL_DIR}" -type f -name "*.gz" 2>/dev/null | wc -l)
        local wal_size=$(du -sh "${WAL_DIR}" 2>/dev/null | cut -f1)
        log "WAL files: ${wal_count} (${wal_size})"

        # Show latest WAL files
        log "Latest WAL files:"
        ls -lht "${WAL_DIR}" 2>/dev/null | head -5 || echo "  No WAL files found"
    else
        log "WAL directory not found: ${WAL_DIR}"
    fi

    # Show base backups
    if [[ -d "${BASE_DIR}" ]]; then
        local base_count=$(find "${BASE_DIR}" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
        local base_size=$(du -sh "${BASE_DIR}" 2>/dev/null | cut -f1)
        log "Base backups: ${base_count} (${base_size})"

        log "Available base backups:"
        for backup in $(ls -1t "${BASE_DIR}" 2>/dev/null); do
            if [[ -f "${BASE_DIR}/${backup}/backup_info.json" ]]; then
                local info=$(cat "${BASE_DIR}/${backup}/backup_info.json")
                local created=$(echo "${info}" | jq -r '.created_at // "unknown"')
                local size=$(echo "${info}" | jq -r '.size_bytes // 0' | numfmt --to=iec 2>/dev/null || echo "unknown")
                log "  ${backup} (created: ${created}, size: ${size})"
            else
                log "  ${backup}"
            fi
        done
    else
        log "Base backup directory not found: ${BASE_DIR}"
    fi

    # Show current WAL position
    local current_wal
    current_wal=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "SELECT pg_walfile_name(pg_current_wal_lsn())" 2>/dev/null | tr -d ' ') || true
    if [[ -n "${current_wal}" ]]; then
        log "Current WAL file: ${current_wal}"
    fi
}

restore_pitr() {
    local TARGET_TIME="${1:-}"
    local BASE_BACKUP="${2:-}"

    log "Point-in-Time Recovery (PITR)"
    log "=============================="

    if [[ -z "${TARGET_TIME}" ]]; then
        echo "Usage: $0 restore-pitr <target_time> [base_backup]"
        echo ""
        echo "Arguments:"
        echo "  target_time  - Recovery target time in ISO format (e.g., '2025-01-20 02:00:00')"
        echo "  base_backup  - Base backup directory name (optional, uses latest if not specified)"
        echo ""
        echo "Example:"
        echo "  $0 restore-pitr '2025-01-20 02:00:00'"
        echo "  $0 restore-pitr '2025-01-20 02:00:00' base_20250119_020000"
        exit 1
    fi

    # Find base backup
    if [[ -z "${BASE_BACKUP}" ]]; then
        BASE_BACKUP=$(ls -1t "${BASE_DIR}" | head -1)
        if [[ -z "${BASE_BACKUP}" ]]; then
            error "No base backups found"
        fi
        log "Using latest base backup: ${BASE_BACKUP}"
    fi

    local BACKUP_PATH="${BASE_DIR}/${BASE_BACKUP}"
    if [[ ! -d "${BACKUP_PATH}" ]]; then
        error "Base backup not found: ${BACKUP_PATH}"
    fi

    log "Target recovery time: ${TARGET_TIME}"
    log "Base backup: ${BACKUP_PATH}"
    log ""
    log "PITR restore commands:"
    log ""
    log "1. Stop PostgreSQL:"
    log "   docker compose stop postgres"
    log ""
    log "2. Backup current data (optional but recommended):"
    log "   docker compose run --rm backup tar -czf /backups/postgres/pre-restore-\$(date +%Y%m%d_%H%M%S).tar.gz /var/lib/postgresql/data"
    log ""
    log "3. Clear PostgreSQL data directory:"
    log "   docker compose run --rm postgres rm -rf /var/lib/postgresql/data/*"
    log ""
    log "4. Restore base backup:"
    log "   docker compose run --rm backup sh -c 'tar -xzf ${BACKUP_PATH}/base.tar.gz -C /var/lib/postgresql/data'"
    log ""
    log "5. Copy WAL files:"
    log "   docker compose run --rm backup cp -r ${WAL_DIR}/* /var/lib/postgresql/data/pg_wal/"
    log ""
    log "6. Create recovery.signal and configure recovery:"
    log "   docker compose run --rm postgres sh -c 'touch /var/lib/postgresql/data/recovery.signal'"
    log "   docker compose run --rm postgres sh -c \"echo \\\"restore_command = 'cp ${WAL_DIR}/%f %p'\\\" >> /var/lib/postgresql/data/postgresql.auto.conf\""
    log "   docker compose run --rm postgres sh -c \"echo \\\"recovery_target_time = '${TARGET_TIME}'\\\" >> /var/lib/postgresql/data/postgresql.auto.conf\""
    log "   docker compose run --rm postgres sh -c \"echo \\\"recovery_target_action = 'promote'\\\" >> /var/lib/postgresql/data/postgresql.auto.conf\""
    log ""
    log "7. Start PostgreSQL:"
    log "   docker compose start postgres"
    log ""
    log "8. Verify recovery and clean up recovery settings:"
    log "   docker compose exec postgres psql -U ${DB_USER} -d ${DB_NAME} -c 'SELECT pg_is_in_recovery()'"
    log "   # After verification, remove recovery settings from postgresql.auto.conf"
}

# Main dispatch
case "${1:-base}" in
    base)
        create_base_backup
        ;;
    wal-status|status)
        show_wal_status
        ;;
    restore-pitr|pitr)
        shift
        restore_pitr "$@"
        ;;
    help|--help|-h)
        echo "PostgreSQL Incremental Backup with WAL Archiving"
        echo ""
        echo "Usage: $0 <command>"
        echo ""
        echo "Commands:"
        echo "  base          Create a new base backup (default)"
        echo "  wal-status    Show WAL archive status"
        echo "  restore-pitr  Show PITR restore instructions"
        echo ""
        echo "Prerequisites:"
        echo "  PostgreSQL must be configured for WAL archiving."
        echo "  Add to postgresql.conf:"
        echo "    wal_level = replica"
        echo "    archive_mode = on"
        echo "    archive_command = 'gzip < %p > ${WAL_DIR}/%f.gz'"
        ;;
    *)
        error "Unknown command: $1. Use --help for usage."
        ;;
esac
