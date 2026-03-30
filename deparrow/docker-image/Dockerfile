# DEparrow Docker Image with DPC Support
# Branded wrapper around bacalhau for Global VM with DPC rewards
#
# Build: docker build -t deparrow/node:latest .
# Run:   docker run -d deparrow/node serve --compute ...

FROM alpine:3.20

LABEL org.opencontainers.image.title="DEparrow Node"
LABEL org.opencontainers.image.description="DEparrow Global Virtual Machine - AI Agents Buy Compute to Run Themselves"
LABEL org.opencontainers.image.vendor="DEparrow"

# Install dependencies
RUN apk add --no-cache ca-certificates docker-cli

# Copy locally built bacalhau with DPC support
COPY bacalhau /usr/local/bin/bacalhau.real
RUN chmod +x /usr/local/bin/bacalhau.real

# Copy branding entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create DEparrow data directory
RUN mkdir -p /root/.bacalhau /root/.deparrow /etc/deparrow
RUN ln -sf /root/.bacalhau /root/.deparrow

# Set DEparrow branding
ENV DEPARROW_NODE=true
ENV DEPARROW_VERSION=1.0.0

# DPC Configuration (can be overridden at runtime)
ENV DPC_ENABLED=false
ENV DPC_RPC_ENDPOINT=http://34.180.51.11:26657
ENV DPC_CHAIN_ID=dpc-testnet-1

# Use our branded entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Default command
CMD ["serve", "--compute"]
