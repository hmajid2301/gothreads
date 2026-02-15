FROM python:3.11-slim

RUN apt-get update && apt-get install -y \
    git libglib2.0-0 libsm6 libxext6 libxrender-dev libgomp1 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install CPU-only PyTorch (smaller image, no CUDA needed)
RUN pip install --no-cache-dir \
    torch==2.1.2 torchvision==0.16.2 \
    --index-url https://download.pytorch.org/whl/cpu

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Clone CatVTON repo
RUN git clone https://github.com/Zheng-Chong/CatVTON.git /app/catvton

COPY app.py .
COPY create_nsfw_placeholder.py .

# Create NSFW placeholder image (required by CatVTON safety checker)
RUN mkdir -p /app/catvton/resource/img && \
    python3 create_nsfw_placeholder.py

# Model weights are downloaded on first request and cached in the
# huggingface-cache Docker volume mounted at /root/.cache/huggingface.
# First try-on will be slower (~5-10 min for download + inference).

EXPOSE 8557

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8557", "--timeout-keep-alive", "600"]
