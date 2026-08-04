FROM python:3.12-slim

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    gcc \
    g++ \
    libsndfile1 \
    && rm -rf /var/lib/apt/lists/*

ENV PIP_DEFAULT_TIMEOUT=600
ENV PIP_RETRIES=15

COPY requirements.txt .

RUN --mount=type=cache,target=/root/.cache/pip \
    pip install --upgrade pip setuptools wheel && \
    pip install --default-timeout=600 --retries=15 llvmlite numba && \
    pip install --default-timeout=600 --retries=15 -r requirements.txt

COPY . .

RUN useradd -m appuser && mkdir -p /app/data && chown -R appuser /app
USER appuser

EXPOSE 8001

CMD ["python", "bot.py"]
