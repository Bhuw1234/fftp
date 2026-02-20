FROM python:3.11-slim

WORKDIR /app

# Install async dependencies (bootstrap-server.py uses aiohttp)
RUN pip install --no-cache-dir \
    aiohttp \
    aiohttp-cors \
    asyncpg \
    pydantic \
    PyJWT \
    uvicorn[standard]

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

# Run with uvicorn (async server for aiohttp)
CMD ["uvicorn", "bootstrap-server:app", "--host", "0.0.0.0", "--port", "8080"]
