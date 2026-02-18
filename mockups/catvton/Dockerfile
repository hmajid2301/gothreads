FROM rocm/pytorch:latest

WORKDIR /app

# System dependencies for CatVTON + DensePose + SCHP
RUN apt-get update && apt-get install -y --no-install-recommends \
    git libglib2.0-0 libgl1 libsm6 libxext6 libxrender1 libgomp1 \
    build-essential python3-dev ninja-build \
    && rm -rf /var/lib/apt/lists/*

# Ensure setuptools is new enough (provides distutils shim for Python 3.12)
RUN pip install --upgrade setuptools wheel

# Install detectron2 from official source — compiles _C extension for Python 3.12.
# CatVTON bundles detectron2 with _C.cpython-39 .so which only works on Python 3.9.
# --no-build-isolation so it finds the pre-installed torch from the base image.
RUN pip install --no-build-isolation \
    'git+https://github.com/facebookresearch/detectron2.git'

# Clone CatVTON repo
RUN git clone https://github.com/Zheng-Chong/CatVTON.git /app/catvton

WORKDIR /app/catvton

# Remove CatVTON's bundled detectron2 (has cpython-39 .so, incompatible with Python 3.12).
# The densepose/ directory is pure Python and stays — it imports from the system detectron2.
RUN rm -rf detectron2

# Install CatVTON requirements, skipping packages already in the base image
# to avoid version conflicts (torch, torchvision, setuptools).
# Also remove accelerate pin — latest diffusers from git HEAD needs newer accelerate
# that has clear_device_cache.
RUN sed -i '/^torch==/d' requirements.txt && \
    sed -i '/^torchvision==/d' requirements.txt && \
    sed -i '/^setuptools==/d' requirements.txt && \
    sed -i '/^accelerate==/d' requirements.txt && \
    pip install --no-cache-dir -r requirements.txt

# Fix gradio_client bug: get_type() crashes when schema is a bool instead of dict.
# This happens with gradio_client==1.3.0 + latest diffusers JSON schemas.
COPY patch_gradio_client.py /tmp/
RUN python /tmp/patch_gradio_client.py

# Create NSFW placeholder (required by safety checker)
COPY create_nsfw_placeholder.py /app/
RUN mkdir -p resource/img && python3 /app/create_nsfw_placeholder.py

# Model weights download on first request (cached in /root/.cache/huggingface volume)

EXPOSE 7860

# Run Gradio app
CMD ["python", "app.py"]
