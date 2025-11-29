#!/bin/bash
# Database backup script for deployment
# Usage: ./scripts/db-backup.sh [backup_name]

set -euo pipefail

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
BACKUP_NAME="${1:-backup-$(date +%Y%m%d-%H%M%S)}"
CONTAINER_NAME="${DB_CONTAINER:-inf-sys-postgres}"
DB_NAME="${POSTGRES_DB:-inf_sys_db}"
DB_USER="${POSTGRES_USER:-postgres}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

log_info "Starting database backup: $BACKUP_NAME"

# Check if container is running
if ! docker ps --format '{{.Names}}' | grep -q "^${CONTAINER_NAME}$"; then
    log_error "PostgreSQL container '$CONTAINER_NAME' is not running"
    exit 1
fi

# Create backup
BACKUP_FILE="${BACKUP_DIR}/${BACKUP_NAME}.sql.gz"

log_info "Creating backup to: $BACKUP_FILE"

docker exec "$CONTAINER_NAME" pg_dump \
    -U "$DB_USER" \
    -d "$DB_NAME" \
    --no-owner \
    --no-acl \
    --clean \
    --if-exists \
    | gzip > "$BACKUP_FILE"

# Verify backup
if [ -f "$BACKUP_FILE" ] && [ -s "$BACKUP_FILE" ]; then
    BACKUP_SIZE=$(du -h "$BACKUP_FILE" | cut -f1)
    log_info "Backup completed successfully"
    log_info "Backup file: $BACKUP_FILE"
    log_info "Backup size: $BACKUP_SIZE"

    # Keep only last 10 backups
    log_info "Cleaning old backups (keeping last 10)..."
    ls -t "$BACKUP_DIR"/*.sql.gz 2>/dev/null | tail -n +11 | xargs -r rm -f

    echo "$BACKUP_FILE"
else
    log_error "Backup failed or file is empty"
    exit 1
fi
