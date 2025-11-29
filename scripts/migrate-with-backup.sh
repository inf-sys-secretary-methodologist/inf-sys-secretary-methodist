#!/bin/bash
# Safe migration script with automatic backup
# Usage: ./scripts/migrate-with-backup.sh [up|down|version]

set -euo pipefail

ACTION="${1:-up}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# Load environment variables
if [ -f "$PROJECT_DIR/.env" ]; then
    export $(grep -v '^#' "$PROJECT_DIR/.env" | xargs)
elif [ -f "$PROJECT_DIR/.env.deploy" ]; then
    export $(grep -v '^#' "$PROJECT_DIR/.env.deploy" | xargs)
else
    log_error "No .env or .env.deploy file found"
    exit 1
fi

DB_HOST="${DB_HOST:-postgres}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${POSTGRES_USER:-postgres}"
DB_PASSWORD="${POSTGRES_PASSWORD:-}"
DB_NAME="${POSTGRES_DB:-inf_sys_db}"
DB_SSL_MODE="${DB_SSL_MODE:-disable}"

DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}"

case "$ACTION" in
    version)
        log_info "Checking current migration version..."
        docker run --rm --network host \
            migrate/migrate:latest \
            -path=/migrations \
            -database "$DATABASE_URL" \
            version
        exit 0
        ;;

    up)
        log_step "1/4 Creating backup before migration..."
        BACKUP_NAME="pre-migration-$(date +%Y%m%d-%H%M%S)"
        if "$SCRIPT_DIR/db-backup.sh" "$BACKUP_NAME"; then
            log_info "Backup created successfully"
        else
            log_error "Backup failed, aborting migration"
            exit 1
        fi

        log_step "2/4 Getting current migration version..."
        CURRENT_VERSION=$(docker run --rm --network host \
            -v "$PROJECT_DIR/migrations:/migrations" \
            migrate/migrate:latest \
            -path=/migrations \
            -database "$DATABASE_URL" \
            version 2>&1 | tail -1 || echo "0")
        log_info "Current version: $CURRENT_VERSION"

        log_step "3/4 Running migrations..."
        if docker run --rm --network host \
            -v "$PROJECT_DIR/migrations:/migrations" \
            migrate/migrate:latest \
            -path=/migrations \
            -database "$DATABASE_URL" \
            up; then
            log_info "Migrations completed successfully"
        else
            log_error "Migration failed!"
            log_warn "You can restore from backup: ./scripts/db-restore.sh ./backups/${BACKUP_NAME}.sql.gz"
            exit 1
        fi

        log_step "4/4 Verifying new version..."
        NEW_VERSION=$(docker run --rm --network host \
            -v "$PROJECT_DIR/migrations:/migrations" \
            migrate/migrate:latest \
            -path=/migrations \
            -database "$DATABASE_URL" \
            version 2>&1 | tail -1 || echo "unknown")
        log_info "New version: $NEW_VERSION"
        log_info "Migration completed! Backup available at: ./backups/${BACKUP_NAME}.sql.gz"
        ;;

    down)
        log_warn "Rolling back last migration..."
        read -p "Are you sure? (yes/no): " CONFIRM
        if [ "$CONFIRM" != "yes" ]; then
            log_info "Rollback cancelled"
            exit 0
        fi

        log_step "1/2 Creating backup before rollback..."
        BACKUP_NAME="pre-rollback-$(date +%Y%m%d-%H%M%S)"
        "$SCRIPT_DIR/db-backup.sh" "$BACKUP_NAME"

        log_step "2/2 Rolling back..."
        docker run --rm --network host \
            -v "$PROJECT_DIR/migrations:/migrations" \
            migrate/migrate:latest \
            -path=/migrations \
            -database "$DATABASE_URL" \
            down 1

        log_info "Rollback completed"
        ;;

    *)
        echo "Usage: $0 [up|down|version]"
        echo ""
        echo "Commands:"
        echo "  up      - Run all pending migrations (with backup)"
        echo "  down    - Rollback last migration (with backup)"
        echo "  version - Show current migration version"
        exit 1
        ;;
esac
