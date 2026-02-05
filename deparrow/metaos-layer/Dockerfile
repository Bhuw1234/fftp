FROM python:3.11-slim

WORKDIR /app

# Install dependencies
RUN pip install --no-cache-up flask flask-cors flask-jwt-extended psycopg2-binary redis gunicorn

# Copy application
COPY bootstrap-server.py .

# Create non-root user
RUN useradd -m -u 1000 deparrow && chown -R deparrow:deparrow /app
USER deparrow

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Run with gunicorn
CMD ["gunicorn", "--bind", "0.0.0.0:8080", "--workers", "4", "bootstrap-server:app"]
