FROM alpine:3.21

# Install required packages
RUN apk add --no-cache \
    bash \
    postgresql16-client \
    curl \
    gzip \
    tar \
    tzdata \
    dcron

# Install MinIO Client (mc) - detect architecture
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64|arm64) MC_ARCH="arm64" ;; \
        x86_64|amd64) MC_ARCH="amd64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    curl -sSL "https://dl.min.io/client/mc/release/linux-${MC_ARCH}/mc" -o /usr/local/bin/mc && \
    chmod +x /usr/local/bin/mc

# Create backup directories
RUN mkdir -p /backups/postgres /backups/minio /scripts

# Copy scripts
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# Create entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set timezone
ENV TZ=Europe/Moscow

ENTRYPOINT ["/entrypoint.sh"]
