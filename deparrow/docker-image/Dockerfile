# DEparrow Docker Image
# Branded wrapper around bacalhau for Global VM
#
# Build: docker build -t deparrow/node:latest .
# Run:   docker run -d deparrow/node serve --compute ...

FROM ghcr.io/bacalhau-project/bacalhau:latest

LABEL org.opencontainers.image.title="DEparrow Node"
LABEL org.opencontainers.image.description="DEparrow Global Virtual Machine - AI Agents Buy Compute to Run Themselves"
LABEL org.opencontainers.image.vendor="DEparrow"

# Copy branding entrypoint
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create DEparrow data directory symlink
RUN ln -sf /root/.bacalhau /root/.deparrow 2>/dev/null || true

# Set DEparrow branding
ENV DEPARROW_NODE=true
ENV DEPARROW_VERSION=1.0.0

# Use our branded entrypoint
ENTRYPOINT ["/entrypoint.sh"]

# Default command
CMD ["serve", "--compute"]
