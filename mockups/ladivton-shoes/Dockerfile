FROM rocm/pytorch:latest

WORKDIR /app

# System dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git libglib2.0-0 libgl1 libsm6 libxext6 libxrender1 \
    && rm -rf /var/lib/apt/lists/*

# Clone LaDI-VTON-Shoes repo
RUN git clone https://github.com/sheep8888/ladi-vton-shoes.git /app/ladi-vton-shoes

# Install Python dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy the API wrapper
COPY app.py .

# Model weights are downloaded on first request via torch.hub
# and cached in the huggingface-cache Docker volume.

EXPOSE 8558

CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8558", "--timeout-keep-alive", "600"]
