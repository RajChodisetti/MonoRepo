FROM python:3.12-slim

WORKDIR /app

# System deps needed by Silero VAD (onnxruntime) and audio processing
RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

# Slow PyPI downloads (llvmlite ~55MB) often hit pip's default 15s read timeout
ENV PIP_DEFAULT_TIMEOUT=600
ENV PIP_RETRIES=15

COPY requirements.txt .

# Cache pip wheels between rebuilds; pre-install heavy numba stack first
RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip setuptools wheel && \
    pip install --default-timeout=600 --retries=15 llvmlite numba && \
    pip install --default-timeout=600 --retries=15 -r requirements.txt

COPY . .

# Do not run as root in production. Ensure /app/data exists before Docker
# initializes a named volume there so call logs remain writable by appuser.
RUN useradd -m appuser && mkdir -p /app/data && chown -R appuser /app
USER appuser

EXPOSE 8000

CMD ["python", "bot.py"]
