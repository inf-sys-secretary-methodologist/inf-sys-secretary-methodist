FROM alpine:3.21

# Install required packages
RUN apk add --no-cache \
    bash \
    postgresql16-client \
    curl \
    gzip \
    tar \
    tzdata \
    dcron \
    # For notifications and metrics
    jq \
    openssl \
    # For encryption (GPG)
    gnupg

# Install age encryption tool
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64|arm64) AGE_ARCH="arm64" ;; \
        x86_64|amd64) AGE_ARCH="amd64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    AGE_VERSION="v1.2.0" && \
    curl -sSL "https://github.com/FiloSottile/age/releases/download/${AGE_VERSION}/age-${AGE_VERSION}-linux-${AGE_ARCH}.tar.gz" | \
    tar -xz -C /tmp && \
    mv /tmp/age/age /usr/local/bin/age && \
    mv /tmp/age/age-keygen /usr/local/bin/age-keygen && \
    chmod +x /usr/local/bin/age /usr/local/bin/age-keygen && \
    rm -rf /tmp/age

# Install MinIO Client (mc) - detect architecture
RUN ARCH=$(uname -m) && \
    case "$ARCH" in \
        aarch64|arm64) MC_ARCH="arm64" ;; \
        x86_64|amd64) MC_ARCH="amd64" ;; \
        *) echo "Unsupported architecture: $ARCH" && exit 1 ;; \
    esac && \
    curl -sSL "https://dl.min.io/client/mc/release/linux-${MC_ARCH}/mc" -o /usr/local/bin/mc && \
    chmod +x /usr/local/bin/mc

# Create directories
RUN mkdir -p \
    /backups/postgres \
    /backups/minio \
    /scripts \
    /var/lib/node_exporter/textfile_collector

# Copy scripts
COPY scripts/ /scripts/
RUN chmod +x /scripts/*.sh

# Create entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Set timezone
ENV TZ=Europe/Moscow

# Default environment variables
ENV METRICS_ENABLED=true \
    NOTIFY_ON_FAILURE=true \
    NOTIFY_ON_SUCCESS=false \
    BACKUP_ENCRYPTION_ENABLED=false \
    BACKUP_ENCRYPTION_TYPE=age

ENTRYPOINT ["/entrypoint.sh"]
