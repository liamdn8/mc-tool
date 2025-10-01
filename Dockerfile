FROM alpine:3.20

# Install basic tools including awscurl for MinIO API calls and performance monitoring
RUN apk add --no-cache \
      bash \
      curl \
      jq \
      zip \
      unzip \
      tar \
      coreutils \
      busybox-extras \
      procps \
      iputils \
      net-tools \
      ca-certificates \
      && update-ca-certificates \
      && rm -rf /var/cache/apk/*

# Download mc (MinIO client) latest release
RUN curl -sSL https://dl.min.io/client/mc/release/linux-amd64/mc \
      -o /usr/local/bin/mc \
    && chmod +x /usr/local/bin/mc

# # Download golang latest release
# RUN curl -sSL https://go.dev/dl/go1.25.1.linux-amd64.tar.gz \
#     -o /tmp/go.tar.gz \
#     && tar -C /usr/local -xzf /tmp/go.tar.gz \
#     && rm /tmp/go.tar.gz

# # Set Go environment variables
# ENV GOPATH=/go \
#     PATH=$PATH:/usr/local/go/bin:/go/bin

# Copy the mc-tool binary
COPY build/mc-tool-portable /usr/local/bin/mc-tool
RUN chmod +x /usr/local/bin/mc-tool

# Create a non-root user to run the tool
RUN adduser -D -u 1000 -h /home/vt_admin vt_admin \
    && mkdir -p /home/vt_admin \
    && chown -R vt_admin:vt_admin /home/vt_admin

WORKDIR /home/vt_admin
USER vt_admin

ENTRYPOINT ["/bin/bash"]